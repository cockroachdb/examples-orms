package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/examples-orms/go/gopg/model"
	"github.com/go-pg/pg/v10"
	"github.com/julienschmidt/httprouter"
)

// Server is an http server that handles REST requests.
type Server struct {
	db *pg.DB
}

// NewServer creates a new instance of a Server.
func NewServer(db *pg.DB) *Server {
	return &Server{db: db}
}

// RegisterRouter registers a router onto the Server.
func (s *Server) RegisterRouter(router *httprouter.Router) {
	router.GET("/ping", s.ping)

	router.GET("/customer", s.getCustomers)
	router.POST("/customer", s.createCustomer)
	router.GET("/customer/:customerID", s.getCustomer)
	router.PUT("/customer/:customerID", s.updateCustomer)
	router.DELETE("/customer/:customerID", s.deleteCustomer)

	router.GET("/product", s.getProducts)
	router.POST("/product", s.createProduct)
	router.GET("/product/:productID", s.getProduct)
	router.PUT("/product/:productID", s.updateProduct)
	router.DELETE("/product/:productID", s.deleteProduct)

	router.GET("/order", s.getOrders)
	router.POST("/order", s.createOrder)
	router.GET("/order/:orderID", s.getOrder)
	router.PUT("/order/:orderID", s.updateOrder)
	router.DELETE("/order/:orderID", s.deleteOrder)
	router.POST("/order/:orderID/product", s.addProductToOrder)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeTextResult(w, "go/gopg")
}

func (s *Server) getCustomers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var customers []model.Customer
	if err := s.db.Model(&customers).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, customers)
	}
}

func (s *Server) createCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var customer model.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	if _, err := s.db.Model(&customer).Insert(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, customer)
	}
}

func (s *Server) getCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	customerID, err := strconv.Atoi(ps.ByName("customerID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	customer := model.Customer{
		ID: customerID,
	}
	if err := s.db.Model(&customer).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, customer)
	}
}

func (s *Server) updateCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var customer model.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	customerID, err := strconv.Atoi(ps.ByName("customerID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	customer.ID = customerID
	if _, err := s.db.Model(&customer).Update(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, customer)
	}
}

func (s *Server) deleteCustomer(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	customerID, err := strconv.Atoi(ps.ByName("customerID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	customer := model.Customer{
		ID: customerID,
	}
	res, err := s.db.Model(&customer).WherePK().Delete()
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else if res.RowsAffected() == 0 {
		http.Error(w, "", http.StatusNotFound)
	} else {
		writeTextResult(w, "ok")
	}
}

func (s *Server) getProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var products []model.Product
	if err := s.db.Model(&products).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, products)
	}
}

func (s *Server) createProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var product model.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	if _, err := s.db.Model(&product).Insert(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, product)
	}
}

func (s *Server) getProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	productID, err := strconv.Atoi(ps.ByName("productID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	product := model.Product{
		ID: productID,
	}
	if err := s.db.Model(&product).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, product)
	}
}

func (s *Server) updateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var product model.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	productID, err := strconv.Atoi(ps.ByName("productID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	product.ID = productID
	if _, err := s.db.Model(&product).Update(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, product)
	}
}

func (s *Server) deleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	productID, err := strconv.Atoi(ps.ByName("productID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	product := model.Product{
		ID: productID,
	}
	res, err := s.db.Model(&product).WherePK().Delete()
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else if res.RowsAffected() == 0 {
		http.Error(w, "", http.StatusNotFound)
	} else {
		writeTextResult(w, "ok")
	}
}

func (s *Server) getOrders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var orders []model.Order
	if err := s.db.Model(&orders).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, orders)
	}
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	if order.Customer.ID == 0 {
		http.Error(w, "must specify user", http.StatusBadRequest)
		return
	}
	if err := s.db.Model(&order.Customer).Where("id = ?", order.Customer.ID).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	for i, product := range order.Products {
		if product.ID == 0 {
			http.Error(w, "must specify a product ID", http.StatusBadRequest)
			return
		}
		if err := s.db.Model(&order.Products[i]).Where("id = ?", product.ID).Select(); err != nil {
			http.Error(w, err.Error(), errToStatusCode(err))
			return
		}
	}

	if _, err := s.db.Model(&order).Insert(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		for _, product := range order.Products {
			var orderProduct model.OrderProduct
			orderProduct.Order = order
			orderProduct.Product = product
			if _, err := s.db.Model(&orderProduct).Insert(); err != nil {
				http.Error(w, err.Error(), errToStatusCode(err))
			}
		}
		writeJSONResult(w, order)
	}
}

func (s *Server) getOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orderID, err := strconv.Atoi(ps.ByName("orderID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	order := model.Order{
		ID: orderID,
	}
	if err := s.db.Model(&order).Select(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, order)
	}
}

func (s *Server) updateOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	orderID, err := strconv.Atoi(ps.ByName("orderID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	order.ID = orderID
	if _, err := s.db.Model(&order).Update(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, order)
	}
}

func (s *Server) deleteOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orderID, err := strconv.Atoi(ps.ByName("orderID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	order := model.Order{
		ID: orderID,
	}
	res, err := s.db.Model(&order).WherePK().Delete()
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else if res.RowsAffected() == 0 {
		http.Error(w, "", http.StatusNotFound)
	} else {
		writeTextResult(w, "ok")
	}
}

func (s *Server) addProductToOrder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tx, err := s.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}

	orderID, err := strconv.Atoi(ps.ByName("orderID"))
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	order := model.Order{
		ID: orderID,
	}
	if err := s.db.Model(&order).Select(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	const productIDParam = "productID"
	productIDString := r.URL.Query().Get(productIDParam)
	if productIDString == "" {
		tx.Rollback()
		writeMissingParamError(w, productIDParam)
		return
	}

	productID, err := strconv.Atoi(productIDString)
	if err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	}
	addedProduct := model.Product{
		ID: productID,
	}
	if err := s.db.Model(&addedProduct).Select(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	order.Products = append(order.Products, addedProduct)
	if _, err := tx.Model(&order).Insert(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}
	orderProduct := model.OrderProduct{
		Order:   order,
		Product: addedProduct,
	}
	if _, err := tx.Model(&orderProduct).Insert(); err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), errToStatusCode(err))
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), errToStatusCode(err))
	} else {
		writeJSONResult(w, order)
	}
}

func writeTextResult(w http.ResponseWriter, res string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, res)
}

func writeJSONResult(w http.ResponseWriter, res interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

func writeMissingParamError(w http.ResponseWriter, paramName string) {
	http.Error(w, fmt.Sprintf("missing query param %q", paramName), http.StatusBadRequest)
}

func errToStatusCode(err error) int {
	switch err {
	case pg.ErrNoRows:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
