package testing

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cockroachdb/examples-orms/go/gorm/model"
)

const (
	customersTable     = "customers"
	ordersTable        = "orders"
	productsTable      = "products"
	orderProductsTable = "order_products"
)

// These need to be variables so that their address can be taken.
var (
	customerName1 = "Billy"

	productName1       = "Ice Cream"
	productPrice1      = "123.4"
	productPrice1Float = 123.4
)

// testDriver holds testing state and provides a suite of test methods that
// incrementally stress ORM functionality.
type testDriver struct {
	db     *sql.DB
	dbName string
	api    apiHandler
}

func (td testDriver) TestGeneratedTables(t *testing.T) {
	exp := []string{
		customersTable,
		orderProductsTable,
		ordersTable,
		productsTable,
	}

	td.queryAndAssert(t, exp, `
SELECT table_name
FROM information_schema.tables 
WHERE table_schema = $1 
ORDER BY 1`, td.dbName)
}

func (td testDriver) TestGeneratedCustomersTableColumns(t *testing.T) {
	exp := []string{"id", "name"}
	td.testGeneratedColumnsForTable(t, customersTable, exp)
}
func (td testDriver) TestGeneratedOrdersTableColumns(t *testing.T) {
	exp := []string{"customer_id", "id", "subtotal"}
	td.testGeneratedColumnsForTable(t, ordersTable, exp)
}
func (td testDriver) TestGeneratedProductsTableColumns(t *testing.T) {
	exp := []string{"id", "name", "price"}
	td.testGeneratedColumnsForTable(t, productsTable, exp)
}
func (td testDriver) TestGeneratedOrderProductsTableColumns(t *testing.T) {
	exp := []string{"order_id", "product_id"}
	td.testGeneratedColumnsForTable(t, orderProductsTable, exp)
}
func (td testDriver) testGeneratedColumnsForTable(t *testing.T, table string, columns []string) {
	td.queryAndAssert(t, columns, `
SELECT column_name
FROM information_schema.columns 
WHERE table_schema = $1 AND table_name = $2 
ORDER BY 1`, td.dbName, table)
}

func (td testDriver) TestCustomersEmpty(t *testing.T) {
	td.testTableEmpty(t, productsTable)
}
func (td testDriver) TestOrdersTableEmpty(t *testing.T) {
	td.testTableEmpty(t, customersTable)
}
func (td testDriver) TestProductsTableEmpty(t *testing.T) {
	td.testTableEmpty(t, ordersTable)
}
func (td testDriver) TestOrderProductsTableEmpty(t *testing.T) {
	td.testTableEmpty(t, orderProductsTable)
}
func (td testDriver) testTableEmpty(t *testing.T, table string) {
	td.queryAndAssert(t, []string{"0"}, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table))
}

func (td testDriver) TestRetrieveCustomerBeforeCreation(t *testing.T) {
	found, err := td.api.queryCustomers()
	if err != nil {
		t.Fatal(err)
	}

	expected := []model.Customer{}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("expecting customers from api before creation to be %v, found %v", expected, found)
	}
}
func (td testDriver) TestRetrieveProductBeforeCreation(t *testing.T) {
	found, err := td.api.queryProducts()
	if err != nil {
		t.Fatal(err)
	}

	expected := []model.Product{}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("expecting products from api before creation to be %v, found %v", expected, found)
	}
}

func (td testDriver) TestCreateCustomer(t *testing.T) {
	if err := td.api.createCustomer(customerName1); err != nil {
		t.Fatalf("error creating customer: %v", err)
	}
	td.queryAndAssert(t, []string{customerName1}, fmt.Sprintf(`SELECT name FROM %s`, customersTable))
}
func (td testDriver) TestCreateProduct(t *testing.T) {
	if err := td.api.createProduct(productName1, productPrice1Float); err != nil {
		t.Fatalf("error creating product: %v", err)
	}
	td.queryAndAssert(t, []string{row(productName1, productPrice1)}, fmt.Sprintf(`SELECT name, price FROM %s`, productsTable))
}

func (td testDriver) TestRetrieveCustomerAfterCreation(t *testing.T) {
	found, err := td.api.queryCustomers()
	if err != nil {
		t.Fatal(err)
	}

	expected := []model.Customer{
		{Name: &customerName1},
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("expecting customers from api after creation to be %v, found %v", expected, found)
	}
}
func (td testDriver) TestRetrieveProductAfterCreation(t *testing.T) {
	found, err := td.api.queryProducts()
	if err != nil {
		t.Fatal(err)
	}

	expected := []model.Product{
		{Name: &productName1, Price: productPrice1Float},
	}
	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("expecting products from api after creation to be %v, found %v", expected, found)
	}
}

func (td testDriver) queryAndAssert(t *testing.T, expected []string, query string, args ...interface{}) {
	rows, err := td.db.Query(query, args...)
	if err != nil {
		t.Fatal(err)
	}

	found, err := rowsToStringSlice(rows)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, found) {
		t.Fatalf("expecting rows for query %q with args %+v to be %v, found %v", query, args, expected, found)
	}
}

func rowsToStringSlice(rows *sql.Rows) ([]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	vals := make([]interface{}, len(cols))
	strs := make([]string, len(cols))
	for i := range vals {
		vals[i] = &strs[i]
	}

	var s []string
	for rows.Next() {
		rows.Scan(vals...)
		s = append(s, strings.Join(strs, ", "))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return s, nil
}

func row(vals ...interface{}) string {
	var b bytes.Buffer
	for i, val := range vals {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", val)
	}
	return b.String()
}
