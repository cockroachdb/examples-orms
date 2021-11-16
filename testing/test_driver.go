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

type testTableNames struct {
	customersTable     string
	ordersTable        string
	productsTable      string
	orderProductsTable string
}

type testColumnNames struct {
	customersColumns      []string
	ordersColumns         []string
	productsColumns       []string
	ordersProductsColumns []string
}

func (tcn testColumnNames) IsEmpty() bool {
	return len(tcn.customersColumns)+len(tcn.ordersColumns)+len(tcn.productsColumns)+len(tcn.ordersProductsColumns) == 0
}

// These need to be variables so that their address can be taken.
var (
	customerName1 = "Billy"

	productName1       = "Ice Cream"
	productPrice1      = "123.40"
	productPrice1Float = 123.40

	defaultTestTableNames = testTableNames{
		customersTable:     "customers",
		ordersTable:        "orders",
		productsTable:      "products",
		orderProductsTable: "order_products",
	}

	defaultTestColumnNames = testColumnNames{
		customersColumns:      []string{"id", "name"},
		ordersColumns:         []string{"customer_id", "id", "subtotal"},
		productsColumns:       []string{"id", "name", "price"},
		ordersProductsColumns: []string{"order_id", "product_id"},
	}

	djangoTestTableNames = testTableNames{
		customersTable:     "cockroach_example_customers",
		ordersTable:        "cockroach_example_orders",
		productsTable:      "cockroach_example_products",
		orderProductsTable: "cockroach_example_orders_product",
	}

	djangoTestColumnNames = testColumnNames{
		customersColumns:      []string{"id", "name"},
		ordersColumns:         []string{"customer_id", "id", "subtotal"},
		productsColumns:       []string{"id", "name", "price"},
		ordersProductsColumns: []string{"id", "orders_id", "products_id"},
	}
)

// parallelTestGroup maps a set of names to test functions, and will run each
// entry as a subtest in parallel by passing its T method to t.Run.
type parallelTestGroup map[string]func(t *testing.T)

func (ptg parallelTestGroup) T(t *testing.T) {
	for name, f := range ptg {
		t.Run(name, func(subT *testing.T) {
			subT.Parallel()
			f(subT)
		})
	}
}

// testDriver holds testing state and provides a suite of test methods that
// incrementally stress ORM functionality.
type testDriver struct {
	db     *sql.DB
	dbName string
	api    apiHandler
	// Holds the expected table names for this test.
	tableNames testTableNames
	// Holds the expected columns for this test.
	columnNames testColumnNames
}

func (td testDriver) TestGeneratedTables(t *testing.T) {
	exp := []string{
		td.tableNames.customersTable,
		td.tableNames.orderProductsTable,
		td.tableNames.ordersTable,
		td.tableNames.productsTable,
	}

	actual := make(map[string]interface{}, len(exp))
	tables := td.query(t, `
SELECT table_name
FROM information_schema.tables
-- support both the legacy and the new information_schema structures. The former returned
-- the string 'def' as the table_catalog value for all rows. The latter returns 'public' as
-- the table_schema value for all user-created tables.
WHERE (table_catalog = 'def' AND table_schema = $1) OR (table_catalog = $1 AND table_schema = 'public')
ORDER BY 1`, td.dbName)
	for i := range tables {
		actual[tables[i]] = nil
	}

	for i := range exp {
		if _, ok := actual[exp[i]]; !ok {
			t.Fatalf("table %s is missing from generated schema", exp[i])
		}
	}
}

func (td testDriver) TestGeneratedCustomersTableColumns(t *testing.T) {
	td.testGeneratedColumnsForTable(t, td.tableNames.customersTable,
		td.columnNames.customersColumns)
}
func (td testDriver) TestGeneratedOrdersTableColumns(t *testing.T) {
	td.testGeneratedColumnsForTable(t, td.tableNames.ordersTable,
		td.columnNames.ordersColumns)
}
func (td testDriver) TestGeneratedProductsTableColumns(t *testing.T) {
	td.testGeneratedColumnsForTable(t, td.tableNames.productsTable,
		td.columnNames.productsColumns)
}
func (td testDriver) TestGeneratedOrderProductsTableColumns(t *testing.T) {
	td.testGeneratedColumnsForTable(t, td.tableNames.orderProductsTable,
		td.columnNames.ordersProductsColumns)
}
func (td testDriver) testGeneratedColumnsForTable(t *testing.T, table string, columns []string) {
	td.queryAndAssert(t, columns, `
SELECT column_name
FROM information_schema.columns
-- see above about supporting both the legacy and the new information_schema structures.
WHERE ((table_catalog = 'def' AND table_schema = $1) OR (table_catalog = $1 AND table_schema = 'public'))
  AND table_name = $2
  AND column_name != 'rowid'
ORDER BY 1`, td.dbName, table)
}

func (td testDriver) TestCustomersEmpty(t *testing.T) {
	td.testTableEmpty(t, td.tableNames.productsTable)
}
func (td testDriver) TestOrdersTableEmpty(t *testing.T) {
	td.testTableEmpty(t, td.tableNames.customersTable)
}
func (td testDriver) TestProductsTableEmpty(t *testing.T) {
	td.testTableEmpty(t, td.tableNames.ordersTable)
}
func (td testDriver) TestOrderProductsTableEmpty(t *testing.T) {
	td.testTableEmpty(t, td.tableNames.orderProductsTable)
}
func (td testDriver) testTableEmpty(t *testing.T, table string) {
	td.queryAndAssert(t, []string{"0"}, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table))
}

func (td testDriver) TestRetrieveCustomersBeforeCreation(t *testing.T) {
	found, err := td.api.queryCustomers()
	if err != nil {
		t.Fatal(err)
	}

	expected1 := []model.Customer{}
	var expected2 []model.Customer
	if !reflect.DeepEqual(expected1, found) && !reflect.DeepEqual(expected2, found) {
		t.Fatalf("expecting customers from api before creation to be %v or %v, found %v", expected1, expected2, found)
	}
}
func (td testDriver) TestRetrieveProductsBeforeCreation(t *testing.T) {
	found, err := td.api.queryProducts()
	if err != nil {
		t.Fatal(err)
	}

	expected1 := []model.Product{}
	var expected2 []model.Product
	if !reflect.DeepEqual(expected1, found) && !reflect.DeepEqual(expected2, found) {
		t.Fatalf("expecting products from api before creation to be %v or %v, found %v", expected1, expected2, found)
	}
}
func (td testDriver) TestRetrieveOrdersBeforeCreation(t *testing.T) {
	found, err := td.api.queryOrders()
	if err != nil {
		t.Fatal(err)
	}

	expected1 := []model.Order{}
	var expected2 []model.Order
	if !reflect.DeepEqual(expected1, found) && !reflect.DeepEqual(expected2, found) {
		t.Fatalf("expecting orders from api before creation to be %v or %v, found %v", expected1, expected2, found)
	}
}

func (td testDriver) TestCreateCustomer(t *testing.T) {
	if err := td.api.createCustomer(customerName1); err != nil {
		t.Fatalf("error creating customer: %v", err)
	}
	td.queryAndAssert(t, []string{customerName1},
		fmt.Sprintf(`SELECT name FROM %s`, td.tableNames.customersTable))
}
func (td testDriver) TestCreateProduct(t *testing.T) {
	if err := td.api.createProduct(productName1, productPrice1Float); err != nil {
		t.Fatalf("error creating product: %v", err)
	}
	td.queryAndAssert(t, []string{row(productName1, productPrice1)},
		fmt.Sprintf(`SELECT name, price FROM %s`, td.tableNames.productsTable))
}

func (td testDriver) TestCreateOrder(t *testing.T) {
	// Get the single customer ID.
	customerIDs, err := td.queryIDs(t, td.tableNames.customersTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(customerIDs) != 1 {
		t.Fatalf("expected a single customer ID, found %v", customerIDs)
	}
	customerID := customerIDs[0]

	// Get the single product.
	productIDs, err := td.queryIDs(t, td.tableNames.productsTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(productIDs) != 1 {
		t.Fatalf("expected a single product ID, found %v", productIDs)
	}
	productID := productIDs[0]

	if err := td.api.createOrder(customerID, productID, productPrice1Float); err != nil {
		t.Fatalf("error creating order: %v", err)
	}
	td.queryAndAssert(t, []string{row(productPrice1)},
		fmt.Sprintf(`SELECT subtotal FROM %s`, td.tableNames.ordersTable))
}

func (td testDriver) TestRetrieveCustomerAfterCreation(t *testing.T) {
	found, err := td.api.queryCustomers()
	if err != nil {
		t.Fatal(err)
	}

	expected := []model.Customer{
		{Name: &customerName1},
	}
	if !reflect.DeepEqual(expected, cleanCustomers(found)) {
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
	if !reflect.DeepEqual(expected, cleanProducts(found)) {
		t.Fatalf("expecting products from api after creation to be %v, found %v", expected, found)
	}
}
func (td testDriver) TestRetrieveOrderAfterCreation(t *testing.T) {
	found, err := td.api.queryOrders()
	if err != nil {
		t.Fatal(err)
	}

	expected := []model.Order{
		{Subtotal: productPrice1Float},
	}
	if !reflect.DeepEqual(expected, cleanOrders(found)) {
		t.Fatalf("expecting orders from api after creation to be %v, found %v", expected, found)
	}
}

func (td testDriver) queryIDs(t *testing.T, table string) ([]int, error) {
	rows, err := td.db.Query(fmt.Sprintf("SELECT id FROM %s", table))
	if err != nil {
		t.Fatal(err)
	}

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return ids, nil
}

func (td testDriver) query(t *testing.T, query string, args ...interface{}) []string {
	rows, err := td.db.Query(query, args...)
	if err != nil {
		t.Fatal(err)
	}

	found, err := rowsToStringSlice(rows)
	if err != nil {
		t.Fatal(err)
	}
	return found
}

func (td testDriver) queryAndAssert(
	t *testing.T, expected []string, query string, args ...interface{},
) {
	found := td.query(t, query, args...)

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
		if err := rows.Scan(vals...); err != nil {
			return nil, err
		}
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
