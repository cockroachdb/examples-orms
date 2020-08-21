package testing

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/cockroachdb/examples-orms/version"
	// Import postgres driver.
	_ "github.com/lib/pq"
)

// application represents a single instance of an application running an ORM and
// exposing an HTTP REST API.
type application struct {
	language string
	orm      string
}

func (app application) name() string {
	return fmt.Sprintf("%s/%s", app.language, app.orm)
}

func (app application) dir() string {
	return fmt.Sprintf("../%s", app.name())
}

func (app application) dbName() string {
	return fmt.Sprintf("company_%s", app.orm)
}

// customURLSchemes contains custom schemes for database URLs that are needed
// for test apps that rely on a custom ORM dialect.
var customURLSchemes = map[application]string{
	{language: "python", orm: "sqlalchemy"}: "cockroachdb",
}

type tenantServer interface {
	NewTenantServer(proxy bool) (testserver.TestServer, error)
}

// newServer creates a new cockroachDB server.
func newServer(t *testing.T, insecure bool) testserver.TestServer {
	t.Helper()
	var ts testserver.TestServer
	var err error
	if insecure {
		ts, err = testserver.NewTestServer()
	} else {
		ts, err = testserver.NewTestServer(testserver.SecureOpt())
	}
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

// newTenant creates a new SQL Tenant pointed at the given TestServer. See
// TestServer.NewTenantServer for more information.
func newTenant(t *testing.T, ts testserver.TestServer) testserver.TestServer {
	t.Helper()
	tenant, err := ts.(tenantServer).NewTenantServer(false /* proxy */)
	if err != nil {
		t.Fatal(err)
	}
	return tenant
}

// startServerWithApplication launches a test database as a subprocess.
func startServerWithApplication(
	t *testing.T, ts testserver.TestServer, app application,
) (*sql.DB, *url.URL, func()) {
	t.Helper()
	serverURL := ts.PGURL()
	if serverURL == nil {
		t.Fatal("url not found")
	}
	pgURL := *serverURL
	pgURL.Path = app.dbName()
	db, err := sql.Open("postgres", pgURL.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.WaitForInit(); err != nil {
		t.Fatal(err)
	}
	// Create the database if it does not exist.
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + app.dbName()); err != nil {
		t.Fatal(err)
	}
	if scheme, ok := customURLSchemes[app]; ok {
		pgURL.Scheme = scheme
	}
	return db, &pgURL, func() {
		_ = db.Close()
		ts.Stop()
	}
}

func getVersionFromDB(t *testing.T, db *sql.DB) *version.Version {
	t.Helper()
	var crdbVersion string
	if err := db.QueryRow(
		`SELECT value FROM crdb_internal.node_build_info where field = 'Version'`,
	).Scan(&crdbVersion); err != nil {
		t.Fatal(err)
	}
	v, err := version.Parse(crdbVersion)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// initORMApp launches an ORM application as a subprocess and returns a
// function that terminates that process.
func initORMApp(app application, dbURL *url.URL) (func() error, error) {
	cmd := exec.Command("make", "start", "-C", app.dir(), "ADDR="+dbURL.String())
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	// make will launch the application in a child process, and this is the most
	// straightforward way to kill all ancestors.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	killCmd := func() error {
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			return err
		}
		// This error is expected.
		if err := cmd.Wait(); err.Error() != "signal: "+syscall.SIGKILL.String() {
			return err
		}

		// Killing a process is not instant. For example, with the Hibernate server,
		// it often takes ~10 seconds for the listen port to become available after
		// this function is called. This is despite the above code that issues a
		// SIGKILL to the process group for the test server.
		for {
			if !(apiHandler{}).canDial() {
				break
			}
			log.Printf("waiting for app server port to become available")
			time.Sleep(time.Second)
		}

		return nil
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("command %s failed to start: args=%s", cmd.Args, err)
	}

	const maxWait = 3 * time.Minute
	const waitDelay = 250 * time.Millisecond

	for waited := time.Duration(0); ; waited += waitDelay {
		if processState := cmd.ProcessState; processState != nil && processState.Exited() {
			return nil, fmt.Errorf("command %s exited: %v", cmd.Args, cmd.Wait())
		}
		if err := (apiHandler{}).ping(app.name()); err != nil {
			if waited > maxWait {
				if err := killCmd(); err != nil {
					log.Printf("failed to kill command %s with PID %d: %s", cmd.Args, cmd.ProcessState.Pid(), err)
				}
				return nil, err
			}
			time.Sleep(waitDelay)
			continue
		}
		return killCmd, nil
	}
}

var minRequiredVersionsByORMName = map[string]struct {
	v       *version.Version
	skipMsg string
}{
	"django": {
		v:       version.MustParse("v19.1.0-alpha"),
		skipMsg: "TestDjango fails on CRDB <=v2.1 due to missing foreign key support.",
	},
	"activerecord": {
		v:       version.MustParse("v19.2.0-alpha"),
		skipMsg: "TestActiveRecord fails on CRDB <=v19.1 due to missing pg_catalog support.",
	},
}

type testInfo struct {
	language, orm string
	tableNames    testTableNames  // defaults to defaultTestTableNames
	columnNames   testColumnNames // defaults to defaultTestColumnNames
	// insecure is set if ORM does not handle secure servers (client certs).
	// In that case, we start an insecure server (and don't test in tenant
	// mode).
	insecure bool
}

func testORM(t *testing.T, info testInfo) {
	if info.tableNames == (testTableNames{}) {
		info.tableNames = defaultTestTableNames
	}
	if info.columnNames.IsEmpty() {
		info.columnNames = defaultTestColumnNames
	}
	app := application{
		language: info.language,
		orm:      info.orm,
	}

	type testCase struct {
		name  string
		db    *sql.DB
		dbURL *url.URL
	}
	var testCases []testCase
	{
		ts := newServer(t, info.insecure)
		db, dbURL, stopDB := startServerWithApplication(t, ts, app)
		defer stopDB()

		crdbVersion := getVersionFromDB(t, db)
		// Check that this ORM can be run with the given cockroach version.
		if info, ok := minRequiredVersionsByORMName[info.orm]; ok {
			if !crdbVersion.AtLeast(info.v) {
				t.Skip(info.skipMsg)
			}
		}

		testCases = []testCase{
			{
				name:  "SystemTenant",
				db:    db,
				dbURL: dbURL,
			},
		}

		// This cockroach version supports creating tenants, add a test case to
		// run a tenant server. We need at least 20.1-18 for everything to work out
		// as the certificate story was reworked immediately before that version
		// was minted.
		var tenantsSupported bool
		if err := db.QueryRow(`
SELECT
	(major = 20 AND minor = 1 AND unstable > 17)
	OR (major = 20 AND minor > 1)
	OR (major > 20)
FROM
	[
		SELECT
			regexp_extract(v, e'^(\\d+)\\.')::INT8 AS major,
			regexp_extract(v, e'^\\d+\\.(\\d+)')::INT8
				AS minor,
			regexp_extract(v, e'^\\d+\\.\\d+-(\\d+)')::INT8
				AS unstable
		FROM
			[SHOW CLUSTER SETTING version] AS t (v)
	];
`,
		).Scan(&tenantsSupported); err != nil {
			t.Fatalf("unable to read cluster version: %s", err)
		}
		if tenantsSupported {
			tenant := newTenant(t, ts)
			db, dbURL, stopDB := startServerWithApplication(t, tenant, app)
			defer stopDB()
			testCases = append(testCases, testCase{
				name:  "RegularTenant",
				db:    db,
				dbURL: dbURL,
			})
		} else {
			t.Logf("not running tenant test case because minimum tenant version check was not satisfied")
		}
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td := testDriver{
				db:          tc.db,
				dbName:      app.dbName(),
				tableNames:  info.tableNames,
				columnNames: info.columnNames,
			}

			t.Run("FirstRun", func(t *testing.T) {
				stopApp, err := initORMApp(app, tc.dbURL)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := stopApp(); err != nil {
						t.Fatal(err)
					}
				}()

				// Test that the correct tables were generated.
				t.Run("GeneratedTables", td.TestGeneratedTables)

				// Test that the correct columns in those tables were generated.
				t.Run("GeneratedColumns", parallelTestGroup{
					"CustomersTable":     td.TestGeneratedCustomersTableColumns,
					"ProductsTable":      td.TestGeneratedProductsTableColumns,
					"OrdersTable":        td.TestGeneratedOrdersTableColumns,
					"OrderProductsTable": td.TestGeneratedOrderProductsTableColumns,
				}.T)

				// Test that the tables begin empty.
				t.Run("EmptyTables", parallelTestGroup{
					"CustomersTable":     td.TestCustomersEmpty,
					"ProductsTable":      td.TestProductsTableEmpty,
					"OrdersTable":        td.TestOrdersTableEmpty,
					"OrderProductsTable": td.TestOrderProductsTableEmpty,
				}.T)

				// Test that the API returns empty sets for each collection.
				t.Run("RetrieveFromAPIBeforeCreation", parallelTestGroup{
					"Customers": td.TestRetrieveCustomersBeforeCreation,
					"Products":  td.TestRetrieveProductsBeforeCreation,
					"Orders":    td.TestRetrieveOrdersBeforeCreation,
				}.T)

				// Test the creation of initial objects.
				t.Run("CreateCustomer", td.TestCreateCustomer)
				t.Run("CreateProduct", td.TestCreateProduct)

				// Test that the API returns what we just created.
				t.Run("RetrieveFromAPIAfterInitialCreation", parallelTestGroup{
					"Customers": td.TestRetrieveCustomerAfterCreation,
					"Products":  td.TestRetrieveProductAfterCreation,
				}.T)

				// Test the creation of dependent objects.
				t.Run("CreateOrder", td.TestCreateOrder)

				// Test that the API returns what we just created.
				t.Run("RetrieveFromAPIAfterDependentCreation", parallelTestGroup{
					"Order": td.TestRetrieveProductAfterCreation,
				}.T)
			})

			t.Run("SecondRun", func(t *testing.T) {
				stopApp, err := initORMApp(app, tc.dbURL)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := stopApp(); err != nil {
						t.Fatal(err)
					}
				}()

				// Test that the API still returns all created objects.
				t.Run("RetrieveFromAPIAfterRestart", parallelTestGroup{
					"Customers": td.TestRetrieveCustomerAfterCreation,
					"Products":  td.TestRetrieveProductAfterCreation,
					"Order":     td.TestRetrieveProductAfterCreation,
				}.T)
			})
		})
	}
}

func TestGORM(t *testing.T) {
	testORM(t, testInfo{language: "go", orm: "gorm"})
}

func TestGOPG(t *testing.T) {
	testORM(t, testInfo{
		language: "go",
		orm:      "gopg",
		// GoPG does not support client certs:
		// https://github.com/go-pg/pg/blob/v10/options.go
		// If we set up a secure deployment and went through the proxy, it would work (or should anyway), but only
		// via the 'database' parameter; GoPG also does not support the 'options' parameter.
		insecure: true,
	})
}

func TestHibernate(t *testing.T) {
	testORM(t, testInfo{
		language: "java",
		orm:      "hibernate",
		// Possibly does not unescape the path correctly:
		// Caused by: java.io.FileNotFoundException:
		//	%2Ftmp%2Fcockroach-testserver913095208%2Fcerts%2Fca.crt (No such file or directory)
		insecure: true,
	})
}

func TestSequelize(t *testing.T) {
	testORM(t, testInfo{
		language: "node",
		orm:      "sequelize",
		// Requires bespoke code to actually use SSL, see:
		// https://github.com/sequelize/sequelize/issues/10015
		insecure: true,
	})
}

func TestSQLAlchemy(t *testing.T) {
	testORM(t, testInfo{
		language: "python",
		orm:      "sqlalchemy",
	})
}

func TestDjango(t *testing.T) {
	testORM(t, testInfo{
		language:    "python",
		orm:         "django",
		tableNames:  djangoTestTableNames,
		columnNames: djangoTestColumnNames,
		// No support for client certs (at least not via the query string).
		// psycopg2.OperationalError: fe_sendauth: no password supplied
		insecure: true,
	})
}

func TestActiveRecord(t *testing.T) {
	testORM(t, testInfo{language: "ruby", orm: "activerecord"})
}

func TestActiveRecord4(t *testing.T) {
	testORM(t, testInfo{language: "ruby", orm: "ar4"})
}
