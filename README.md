# Cockroach ORM examples

This repo contains example uses of CockroachDB with popular ORMs.
Each example will implement the [sample application](#sample-app)
spec presented below.

See the [CockroachDB ORM Compatibility Plan](https://docs.google.com/a/cockroachlabs.com/spreadsheets/d/17A0EflPqI9yhargK0n4tSw2WogQuVc5YeK-VFmKvXHM/edit?usp=sharing)
for a roadmap towards supporting various ORMs.

## Installation

Clone this repo into your `$GOPATH` manually, e.g.,

```bash
$ cd ~/go/src/github.com/cockroachdb
$ git clone https://github.com/cockroachdb/examples-orms
```

This is required because this project uses Go to drive the automated tests, so it will look for things in your `$GOPATH`.  If you try to clone it to a non-`$GOPATH` directory, it will fail roughly as follows:

```bash
$ cd ~/some/random/dir/examples-orms
$ make test
go test -v -i ./testing
testing/api_handler.go:13:2: cannot find package "github.com/cockroachdb/examples-orms/go/gorm/model" in any of:
	/usr/local/Cellar/go/1.10/libexec/src/github.com/cockroachdb/examples-orms/go/gorm/model (from $GOROOT)
	/Users/rloveland/go/src/github.com/cockroachdb/examples-orms/go/gorm/model (from $GOPATH)
testing/api_handler.go:11:2: cannot find package "github.com/pkg/errors" in any of:
	/usr/local/Cellar/go/1.10/libexec/src/github.com/pkg/errors (from $GOROOT)
	/Users/rloveland/go/src/github.com/pkg/errors (from $GOPATH)
make: *** [test] Error 1
```

However, this is not actually a Go project, so `go get -d` will also fail (hence the need to manually clone).

```bash
$ go get -d github.com/cockroachdb/examples-orms
package github.com/cockroachdb/examples-orms: no Go files in /Users/rloveland/go/src/github.com/cockroachdb/examples-orms
```

## Testing

To run automated testing against all ORMs using the latest binary of CockroachDB, run:

```bash
$ make test
```

To run automated testing against all ORMs using a custom CockroachDB binary, run:

```bash
$ make test COCKROACH_BINARY=/path/to/binary/cockroach
```

To run automated testing against a specific ORM, you can also specify the name with:

```bash
$ make test COCKROACH_BINARY=/path/to/binary/cockroach TESTS=TestSequelize/password
```

These tests require dependencies to be installed on your system. You can install them with:

```bash
$ make deps
```

While running tests locally, you may find it useful to comment out the lines in the Makefile that
install the dependencies for a tool that you don't want to test.

A final option is to run using docker, so that a reproducible build environment is used.

```bash
$ make dockertest COCKROACH_BINARY=/path/to/binary/cockroach
```

## Project Structure

The repository contains a set of directories named after programming
languages. Beneath these language directories are sub-directories
named after specific ORM used for example application implementations.

Each ORM example uses whatever build tool is standard for the language,
but provides a standardized Makefile with a `start` rule, which will
start an instance of the sample application.

For instance, the directory structure for an example application of the
Hibernate ORM will look like:

```
java
└── hibernate
    ├── Makefile
    └── example_source
```

## Sample App

The sample application which each example implements is a [JSON REST API](#json-api)
modeling a **company** with **customers, products, and orders**. The API
exposes access to the management of this company.

The purpose of the example application is to test common data access patterns
so that we can stress various features and implementation details of each
language/ORM.

### Schema

An ideal schema diagram for the sample application looks like:

```
Customer
  |
  v
Order  <->  Product
```

The schema is implemented by each application using ORM-specific constructs to look as
close as possible to:

```sql
CREATE DATABASE IF NOT EXISTS company_{language}_{ORM};

CREATE TABLE IF NOT EXISTS customers (
  id           SERIAL PRIMARY KEY,
  name         STRING
);

CREATE TABLE IF NOT EXISTS orders (
  id           SERIAL PRIMARY KEY,
  subtotal     DECIMAL(18,2)
  customer_id  INT    REFERENCES customers(id),
);

CREATE TABLE IF NOT EXISTS products (
  id           SERIAL PRIMARY KEY,
  name         STRING,
  price        DECIMAL(18,2)
);

CREATE TABLE IF NOT EXISTS product_orders (
  id           SERIAL PRIMARY KEY,
  product_id   INT    REFERENCES products(id),
  order_id     INT    REFERENCES orders(id) ON DELETE CASCADE
);
```

### JSON API

Each example will expose a RESTful JSON API. The endpoints and example `curl`
command lines are:

```
GET    /customer
    curl http://localhost:6543/customer

GET    /customer/:id
    curl http://localhost:6543/customer/1

POST   /customer
    curl -X POST -d '{"id": 1, "name": "bob"}' http://localhost:6543/customer

PUT    /customer/:id
    curl -X PUT -d '{"id": 2, "name": "robert"}' http://localhost:6543/customer/1

DELETE /customer
    curl -X DELETE http://localhost:6543/customer/1

GET    /product
    curl http://localhost:6543/product

GET    /product/:id
    curl http://localhost:6543/product/1

POST   /product
    curl -X POST -d '{"id": 1, "name": "apple", "price": 0.30}' http://localhost:6543/product

PUT    /product
DELETE /product

GET    /order
    curl http://localhost:6543/order

GET    /order/:id
    curl http://localhost:6543/order/1

POST   /order
    curl -X POST -d '{"id": 1, "subtotal": 18.2, "customer": {"id": 1}}' http://localhost:6543/order

PUT    /order
DELETE /order
```

The semantics of each endpoint will be fleshed out when necessary.

## Unresolved Questions

- Can the schema be completely standardized across ORMs without too
  much of a hassle with overriding default type and naming conventions?
