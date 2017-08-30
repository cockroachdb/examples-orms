# README
This is a sample Rails 4 app that uses CockroachDB. It implements the
REST API documented in the toplevel README of this repository.

To run the database migrations and app, ensure you have a CockroachDB instance
running on localhost:26257. Then, do the following:

```
bundle install
rake db:create
rake db:migrate
rails server
curl http://localhost:3000/customer
```
