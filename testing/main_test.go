package testing

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/testserver"

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

// initTestDatabase launches a test database as a subprocess.
func initTestDatabase(t *testing.T, app application) (*sql.DB, *url.URL, func()) {
	ts, err := testserver.NewTestServer()
	if err != nil {
		t.Fatal(err)
	}

	err = ts.Start()
	if err != nil {
		t.Fatal(err)
	}

	url := ts.PGURL()
	if url == nil {
		t.Fatalf("url not found")
	}

	db, err := sql.Open("postgres", url.String())
	if err != nil {
		t.Fatal(err)
	}

	ts.WaitForInit(db)

	// Create the database if it does not exist.
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + app.dbName()); err != nil {
		t.Fatal(err)
	}

	// Connect to the database again, now with the database in the URL.
	url.Path = app.dbName()
	db, err = sql.Open("postgres", url.String())
	if err != nil {
		t.Fatal(err)
	}

	return db, url, func() {
		_ = db.Close()
		ts.Stop()
	}
}

type killFunc func()
type restartFunc func() (killFunc, restartFunc)

// initORMApp launches an ORM application as a subprocess.
func initORMApp(t *testing.T, app application, dbURL *url.URL) (killFunc, restartFunc) {
	addrFlag := fmt.Sprintf("ADDR=%s", dbURL.String())
	args := []string{"make", "start", "-C", app.dir(), addrFlag}

	cmd := exec.Command(args[0], args[1:]...)

	// Set up stderr to display to console and store in a buffer, so we can later
	// verify that it's clean.
	stderr := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(stderr, os.Stderr)

	// make will launch the application in a child process, and this is the most
	// straightforward way to kill all ancestors.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	killed := false
	killCmd := func() {
		if !killed {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			waitForAppExit()
		}
		killed = true
	}

	if err := cmd.Start(); err != nil {
		killCmd()
		t.Fatal(err)
	}
	if cmd.Process != nil {
		log.Printf("process %d started: %s", cmd.Process.Pid, strings.Join(args, " "))
	}

	if err := waitForInit(app); err != nil {
		killCmd()
		t.Fatalf("error waiting for http server initialization: %v stderr=%s", err, stderr.String())
	}

	restartCmd := func() (killFunc, restartFunc) {
		killCmd()
		return initORMApp(t, app, dbURL)
	}

	return killCmd, restartCmd
}

// waitForInit retries until a connection is successfully established.
func waitForInit(app application) error {
	const maxWait = 3 * time.Minute
	const waitDelay = 250 * time.Millisecond
	const maxWaitLoops = int(maxWait / waitDelay)

	var err error
	var api apiHandler
	for i := 0; i < maxWaitLoops; i++ {
		if err = api.ping(app.name()); err == nil {
			return err
		}
		log.Printf("waitForInit: %v", err)
		time.Sleep(waitDelay)
	}
	return err
}

// waitForExit waits indefinitely for the HTTP port of the ORM to stop
// listening.
func waitForAppExit() {
	const waitDelay = time.Second
	var api apiHandler
	for {
		if !api.canDial() {
			break
		}
		log.Print("waiting for app to exit")
		time.Sleep(waitDelay)
	}
}

func testORM(t *testing.T, language, orm string) {
	app := application{
		language: language,
		orm:      orm,
	}

	db, dbURL, stopDB := initTestDatabase(t, app)
	defer stopDB()

	stopApp, restartApp := initORMApp(t, app, dbURL)
	defer stopApp()

	td := testDriver{
		db:     db,
		dbName: app.dbName(),
	}

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

	// Restart the application.
	stopApp, restartApp = restartApp()
	defer stopApp()

	// Test that the API still returns all created objects.
	t.Run("RetrieveFromAPIAfterRestart", parallelTestGroup{
		"Customers": td.TestRetrieveCustomerAfterCreation,
		"Products":  td.TestRetrieveProductAfterCreation,
		"Order":     td.TestRetrieveProductAfterCreation,
	}.T)
}

func TestGORM(t *testing.T) {
	testORM(t, "go", "gorm")
}

func TestHibernate(t *testing.T) {
	testORM(t, "java", "hibernate")
}

func TestActiveRecord(t *testing.T) {
	testORM(t, "ruby", "activerecord")
}
