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
func newServer(t *testing.T, auth authMode) testserver.TestServer {
	t.Helper()
	var ts testserver.TestServer
	var err error
	switch auth {
	case authClientCert:
		ts, err = testserver.NewTestServer(testserver.SecureOpt(), testserver.NonStableDbOpt())
	case authPassword:
		ts, err = testserver.NewTestServer(testserver.SecureOpt(), testserver.RootPasswordOpt("hunter2"), testserver.NonStableDbOpt())
	case authInsecure:
		ts, err = testserver.NewTestServer(testserver.NonStableDbOpt())
	default:
		err = fmt.Errorf("unknown authMode %d", auth)
	}
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

// newTenant creates a new SQL Tenant pointed at the given TestServer. See
// TestServer.NewTenantServer for more information.
func newTenant(t *testing.T, ts testserver.TestServer, proxy bool) testserver.TestServer {
	t.Helper()
	tenant, err := ts.(tenantServer).NewTenantServer(proxy)
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
		v:       version.MustParse("v20.2.0-alpha"),
		skipMsg: "TestDjango fails on CRDB <=v20.1 due to changes in SHOW TABLES.",
	},
	"activerecord": {
		v:       version.MustParse("v19.2.0-alpha"),
		skipMsg: "TestActiveRecord fails on CRDB <=v19.1 due to missing pg_catalog support.",
	},
}

type authMode byte

const (
	// Use client certs. When testing tenants, does not use the proxy (as the proxy does not support client certs).
	authClientCert authMode = iota
	// Use password auth. When testing tenants, tests through a proxy.
	authPassword
	// Use --insecure. When testing tenants, does not use the proxy (as the proxy does not support insecure connections).
	authInsecure

	authModeSentinel // sentinel to iterate over all modes
)

func (mode authMode) String() string {
	switch mode {
	case authClientCert:
		return "client-cert"
	case authPassword:
		return "password"
	case authInsecure:
		return "insecure"
	default:
		return "unknown"
	}
}

type testInfo struct {
	language, orm string
	tableNames    testTableNames  // defaults to defaultTestTableNames
	columnNames   testColumnNames // defaults to defaultTestColumnNames
}

func testORM(t *testing.T, info testInfo, auth authMode) {
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
		ts := newServer(t, auth)
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

		// If the cockroach version supports creating tenants, add a test case to
		// run a tenant server. We need at least v21.2 for everything to work.
		tenantsSupported := crdbVersion.AtLeast(version.MustParse("v21.2.0-alpha"))
		if tenantsSupported {
			// Connect to the tenant through the SQL proxy, which is only supported
			// when using secure+password auth. (The proxy does not support client
			// certs or insecure connections).
			proxySupported := auth == authPassword
			name := "RegularTenant"
			if proxySupported {
				name += "ThroughProxy"
			}
			tenant := newTenant(t, ts, proxySupported)
			db, dbURL, stopDB := startServerWithApplication(t, tenant, app)
			defer stopDB()
			testCases = append(testCases, testCase{
				name:  name,
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

func testORMForAuthModesExcept(t *testing.T, info testInfo, skips map[authMode]string /* mode -> reason */) {
	for auth := authMode(0); auth < authModeSentinel; auth++ {
		t.Run(fmt.Sprint(auth), func(t *testing.T) {
			if msg := skips[auth]; msg != "" {
				t.Skip(msg)
			}
			testORM(t, info, auth)
		})
	}
}

func nothingSkipped() map[authMode]string { return nil }

func TestGORM(t *testing.T) {
	testORMForAuthModesExcept(t, testInfo{language: "go", orm: "gorm"}, nothingSkipped())
}

func TestGOPG(t *testing.T) {
	testORMForAuthModesExcept(t,
		testInfo{language: "go", orm: "gopg"},
		map[authMode]string{
			// https://github.com/go-pg/pg/blob/v10/options.go
			// If we set up a secure deployment and went through the proxy, it would work (or should anyway), but only
			// via the 'database' parameter; GoPG also does not support the 'options' parameter.
			//
			// pg: options other than 'sslmode', 'application_name' and 'connect_timeout' are not supported
			authClientCert: "GoPG does not support custom root cert",
			authPassword:   "GoPG does not support custom root cert",
		})
}

func TestHibernate(t *testing.T) {
	testORMForAuthModesExcept(t, testInfo{language: "java", orm: "hibernate"}, nothingSkipped())
}

func TestSequelize(t *testing.T) {
	testORMForAuthModesExcept(t, testInfo{language: "node", orm: "sequelize"}, nothingSkipped())
}

func TestSQLAlchemy(t *testing.T) {
	testORMForAuthModesExcept(t, testInfo{language: "python", orm: "sqlalchemy"}, nothingSkipped())
}

func TestDjango(t *testing.T) {
	testORMForAuthModesExcept(
		t,
		testInfo{
			language:    "python",
			orm:         "django",
			tableNames:  djangoTestTableNames,
			columnNames: djangoTestColumnNames,
		}, nothingSkipped(),
	)
}

func TestActiveRecord(t *testing.T) {
	testORMForAuthModesExcept(t, testInfo{language: "ruby", orm: "activerecord"}, nothingSkipped())
}
