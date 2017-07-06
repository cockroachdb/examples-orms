'use strict';

var fs        = require('fs');
var Sequelize = require('sequelize-cockroachdb');
var sequelize = new Sequelize(process.env.ADDR, {});
var DataTypes = Sequelize.DataTypes;

if (!Sequelize.supportsCockroachDB) {
  throw new Error("CockroachDB dialect for Sequelize not installed");
}

if (process.env.ADDR === undefined) {
  throw new Error("ADDR (database URL) must be specified.");
}

module.exports.Customer = sequelize.define('customer', {
  name: DataTypes.STRING
}, {
  timestamps: false
});

module.exports.customerToJSON = function(customer) {
  return {
    id: parseInt(this.id),
    name: this.name
  };
};

module.exports.Order = sequelize.define('order', {
  customer_id: DataTypes.INTEGER,
  subtotal: DataTypes.DECIMAL(18, 2)
}, {
  timestamps: false
});

module.exports.orderToJSON = function(order) {
  return{
    id: parseInt(order.id),
    subtotal: order.subtotal,
    customer: {
      id: order.customer_id
    }
  };
};

module.exports.OrderProduct = sequelize.define('order_products', {
  order_id: {
    type: DataTypes.INTEGER,
    primaryKey: true
  },
  product_id: {
    type: DataTypes.INTEGER,
    primaryKey: true
  }
}, {
  timestamps: false
});

module.exports.Product = sequelize.define('product', {
  name: DataTypes.STRING,
  price: DataTypes.DECIMAL(18, 2)
}, {
  timestamps: false
});

module.exports.productToJSON = function(product) {
  return {
    id: parseInt(product.id),
    name: product.name,
    price: product.price,
  };
};

module.exports.sequelize = sequelize;
module.exports.Sequelize = Sequelize;
