package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cockroachdb/examples-orms/go/gopg/model"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/julienschmidt/httprouter"
)

var (
	addr = flag.String("addr", "postgresql://root@localhost:26257/company_gopg?sslmode=disable", "the address of the database")
)

func main() {
	flag.Parse()

	db := setupDB(*addr)
	defer db.Close()

	router := httprouter.New()

	server := NewServer(db)
	server.RegisterRouter(router)

	log.Fatal(http.ListenAndServe(":6543", router))
}

func setupDB(addr string) *pg.DB {
	opt, err := pg.ParseURL(addr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse addr URL %s: %v", addr, err))
	}
	db := pg.Connect(opt)

	// Need to register OrderProduct before creating it because Order references
	// it.
	orm.RegisterTable((*model.OrderProduct)(nil))

	for _, model := range []interface{}{
		(*model.Customer)(nil),
		(*model.Order)(nil),
		(*model.Product)(nil),
		(*model.OrderProduct)(nil),
	} {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			panic(fmt.Sprintf("failed to create a table: %v", err))
		}
	}
	return db
}
