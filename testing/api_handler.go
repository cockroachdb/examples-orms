package testing

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/examples-orms/go/gorm/model"
)

const (
	applicationAddr = "http://localhost:6543"

	pingPath      = applicationAddr + "/ping"
	customersPath = applicationAddr + "/customer"
	ordersPath    = applicationAddr + "/order"
	productsPath  = applicationAddr + "/product"

	jsonContentType = "application/json"
)

// apiHandler takes care of communicating with the application api. It uses GORM's models
// for convenient JSON marshalling/unmarshalling, but this format should be the same
// across all ORMs.
type apiHandler struct{}

func (apiHandler) ping() error {
	_, err := http.Get(pingPath)
	return err
}

func (apiHandler) queryCustomers() ([]model.Customer, error) {
	var customers []model.Customer
	if err := getJSON(customersPath, &customers); err != nil {
		return nil, err
	}
	return customers, nil
}
func (apiHandler) queryProducts() ([]model.Product, error) {
	var products []model.Product
	if err := getJSON(productsPath, &products); err != nil {
		return nil, err
	}
	return products, nil
}
func (apiHandler) queryOrders() ([]model.Order, error) {
	var orders []model.Order
	if err := getJSON(ordersPath, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (apiHandler) createCustomer(name string) error {
	customer := model.Customer{Name: &name}
	return postJSONData(customersPath, customer)
}
func (apiHandler) createProduct(name string, price float64) error {
	product := model.Product{Name: &name, Price: price}
	return postJSONData(productsPath, product)
}
func (apiHandler) createOrder(customerID, productID int, subtotal float64) error {
	order := model.Order{
		Customer: model.Customer{ID: customerID},
		Products: []model.Product{{ID: productID}},
		Subtotal: subtotal,
	}
	return postJSONData(ordersPath, order)
}

func getJSON(path string, result interface{}) error {
	resp, err := http.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func postJSONData(path string, body interface{}) error {
	var bodyBuf bytes.Buffer
	if err := json.NewEncoder(&bodyBuf).Encode(body); err != nil {
		return err
	}

	_, err := http.Post(path, jsonContentType, &bodyBuf)
	return err
}

// These functions clean any non-deterministic fields, such as IDs that are
// generated upon row creation.
func cleanCustomers(customers []model.Customer) []model.Customer {
	for i := range customers {
		customers[i].ID = 0
	}
	return customers
}
func cleanProducts(products []model.Product) []model.Product {
	for i := range products {
		products[i].ID = 0
	}
	return products
}
func cleanOrders(orders []model.Order) []model.Order {
	for i := range orders {
		orders[i].ID = 0
		orders[i].Customer = model.Customer{}
		orders[i].CustomerID = 0
		orders[i].Products = nil
	}
	return orders
}
