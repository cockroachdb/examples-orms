package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cockroachdb/examples-orms/go/gorm/model"
	"github.com/julienschmidt/httprouter"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	addr = flag.String("addr", "postgresql://root@localhost:26257/company_gorm?sslmode=disable", "the address of the database")
)

func main() {
	flag.Parse()

	db := setupDB(*addr)

	router := httprouter.New()

	server := NewServer(db)
	server.RegisterRouter(router)

	log.Fatal(http.ListenAndServe(":6543", router))
}

func setupDB(addr string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(addr))
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Migrate the schema
	if err := db.AutoMigrate(&model.Customer{}, &model.Order{}, &model.Product{}); err != nil {
		panic(err)
	}

	return db
}
