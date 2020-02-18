drop table if exists order_products;
drop table if exists orders;
drop table if exists customers;
drop table if exists products;

create sequence customers_id;
create sequence products_id;
create sequence orders_id;

create table customers (
  id bigint default nextval('customers_id') not null,
  name text not null,

  constraint pk_customers primary key (id)
);

create table products (
  id bigint default nextval('products_id') not null,
  name text not null,
  price numeric(18, 2) not null,

  constraint pk_products primary key (id)
);

create table orders (
  id bigint default nextval('orders_id') not null,
  subtotal numeric(18, 2) not null,
  customer_id bigint not null,

  constraint pk_orders primary key (id),
  constraint fk_customer foreign key (customer_id) references customers
);

create table order_products (
  order_id bigint not null,
  product_id bigint not null,
  quantity bigint not null,

  constraint pk_order_products primary key (order_id, product_id),
  constraint fk_orders foreign key (order_id) references orders,
  constraint fk_products foreign key (product_id) references products
);


-- Some sample data generation:
insert into customers (name)
  values ('John Smith'), ('Jane Doe');

insert into products (name, price)
  values ('Spoon', 2.50), ('Fork', 1.50);

insert into orders (subtotal, customer_id)
  values (6.50, (select id from customers where name = 'John Smith'));

insert into order_products (order_id, product_id, quantity)
  values
    ((select id from orders where subtotal = 6.50), (select id from products where name = 'Spoon'), 2),
    ((select id from orders where subtotal = 6.50), (select id from products where name = 'Fork'), 1)