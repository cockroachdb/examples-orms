package testing

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/testserver"

	// Import postgres driver.
	_ "github.com/lib/pq"
)

// application represents a single instance of a application running an ORM and
// exposing an HTTP REST API.
type application struct {
	language string
	orm      string
}

func (app application) dir() string {
	return fmt.Sprintf("../%s/%s", app.language, app.orm)
}

func (app application) dbName() string {
	return fmt.Sprintf("company_%s", app.orm)
}

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

	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + app.dbName()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("SET DATABASE = " + app.dbName()); err != nil {
		t.Fatal(err)
	}
	url.Path = app.dbName()

	return db, url, func() {
		_ = db.Close()
		ts.Stop()
	}
}

func initORMApp(t *testing.T, app application, dbURL *url.URL) func() {
	addrFlag := fmt.Sprintf("ADDR=%s", dbURL.String())
	args := []string{"make", "start", "-C", app.dir(), addrFlag}

	cmd := exec.Command(args[0], args[1:]...)

	// make will launch the application in a child process, and this is the most
	// straightforward way to kill all ancestors.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	killCmd := func() {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	// Set up stderr so we can later verify that it's clean.
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if cmd.Process != nil {
		log.Printf("process %d started: %s", cmd.Process.Pid, strings.Join(args, " "))
	}

	time.Sleep(3 * time.Second)
	if s := stderr.String(); len(s) > 0 {
		t.Fatalf("stderr=%s", s)
	}

	return killCmd
}

func testORM(t *testing.T, language, orm string) {
	app := application{
		language: language,
		orm:      orm,
	}

	db, dbURL, stopDB := initTestDatabase(t, app)
	defer stopDB()

	stopApp := initORMApp(t, app, dbURL)
	defer stopApp()

	td := testDriver{
		db:     db,
		dbName: app.dbName(),
	}

	// Test that the correct tables were generated.
	t.Run("TestGeneratedTables", td.TestGeneratedTables)

	// Test that the correct columns in those tables were generated.
	t.Run("TestGeneratedCustomersTableColumns", td.TestGeneratedCustomersTableColumns)
	t.Run("TestGeneratedOrdersTableColumns", td.TestGeneratedOrdersTableColumns)
	t.Run("TestGeneratedProductsTableColumns", td.TestGeneratedProductsTableColumns)
	t.Run("TestGeneratedOrderProductsTableColumns", td.TestGeneratedOrderProductsTableColumns)

	// Test that the tables begin empty.
	t.Run("TestOrdersTableEmpty", td.TestOrdersTableEmpty)
	t.Run("TestProductsTableEmpty", td.TestProductsTableEmpty)
	t.Run("TestCustomersEmpty", td.TestCustomersEmpty)
	t.Run("TestOrderProductsTableEmpty", td.TestOrderProductsTableEmpty)

	// Test the creation of objects.
	t.Run("TestRetrieveCustomerBeforeCreation", td.TestRetrieveCustomerBeforeCreation)
	t.Run("TestRetrieveProductBeforeCreation", td.TestRetrieveProductBeforeCreation)
	t.Run("TestCreateCustomer", td.TestCreateCustomer)
	t.Run("TestCreateProduct", td.TestCreateProduct)
	t.Run("TestRetrieveCustomerAfterCreation", td.TestRetrieveCustomerAfterCreation)
	t.Run("TestRetrieveProductAfterCreation", td.TestRetrieveProductAfterCreation)
}

func TestGORM(t *testing.T) {
	testORM(t, "go", "gorm")
}
