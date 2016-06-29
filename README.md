# Cockroach ORM examples

This repo contains example uses of CockroachDB with popular ORMs.
Each example will implement the [sample application](#sample-app) 
spec presented below. 

See the [CockroachDB ORM Compatibility Plan](https://docs.google.com/a/cockroachlabs.com/spreadsheets/d/17A0EflPqI9yhargK0n4tSw2WogQuVc5YeK-VFmKvXHM/edit?usp=sharing)
for a roadmap towards supporting various ORMs.

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

A schema diagram for the sample application looks like:

```
Customer
  |
  v
Order  <->  Product 
```

The schema is implemented to look as close as possible to:

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

Each example will expose a RESTful JSON API with the following endpoints:

```
GET    /customer
GET    /customer/:id
POST   /customer
PUT    /customer
DELETE /customer

GET    /product
GET    /product/:id
POST   /product
PUT    /product
DELETE /product

GET    /order
GET    /order/:id
POST   /order
PUT    /order
DELETE /order
```

The semantics of each endpoint will be fleshed out when necessary.

## Testing

_Automated testing does not have to be included in the initial development of this repo._

To test each application, a CockroachDB instance will be spun up, and 
each test example will be run against it in a Docker container using `make start`.
A series of HTTP requests with an expected output will be issued against each example,
and the results will be verified.

## Unresolved Questions

- Can the schema be completely standardized across ORMs without too
  much of a hassle with overriding default type and naming conventions?
