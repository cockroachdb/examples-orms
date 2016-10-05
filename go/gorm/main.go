package main

import (
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/julienschmidt/httprouter"
)

func main() {
	db := setupDB()
	defer db.Close()

	router := httprouter.New()

	server := NewServer(db)
	server.RegisterRouter(router)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupDB() *gorm.DB {
	db, err := gorm.Open("postgres", "postgresql://root@localhost:26257/company_gorm?sslmode=disable")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	migrateDB(db)

	// Initialize the database if it's empty.
	var count int
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		// Create Products.
		p1 := "P1"
		p2 := "P2"
		db.Create(&Product{Name: &p1, Price: 22.2})
		db.Create(&Product{Name: &p2, Price: 2.2})

		// Create Customers.
		john := "John"
		fred := "Fred"
		db.Create(&Customer{Name: &john})
		db.Create(&Customer{Name: &fred})

		// Create an Order.
		{
			tx := db.Begin()

			var product Product
			tx.First(&product, "name = ?", "P2")

			var customer Customer
			tx.First(&customer, "name = ?", "Fred")

			tx.Create(&Order{
				Subtotal: product.Price,
				Customer: customer,
				Products: []Product{product},
			})
			tx.Commit()
		}
	}

	return db
}
