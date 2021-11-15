'use strict';

var fs        = require('fs');
var Sequelize = require('sequelize-cockroachdb');

if (process.env.ADDR === undefined) {
  throw new Error("ADDR (database URL) must be specified.");
}

var url = new URL(process.env.ADDR);
var opts = {
  dialect: "postgres",
  username: url.username,
  host: url.hostname,
  port: url.port,
  database: url.pathname.substring(1), // ignore leading '/'
  dialectOptions: {
    cockroachdbTelemetryDisabled: true,
    ssl: {},
  },
  logging: false,
};

if (url.password) {
  opts.password = url.password
}
if (url.searchParams.has("options")) {
  var pgOpts = url.searchParams.get("options")
  var cluster = pgOpts.match(/cluster=([^\s]+)/)[1]
  opts.database = `${cluster}.${opts.database}`
}
if (url.searchParams.get("sslmode") === "disable") {
  delete opts.dialectOptions.ssl
} else {
  if (url.searchParams.has("sslrootcert")) {
    opts.dialectOptions.ssl.ca = fs.readFileSync(url.searchParams.get("sslrootcert").toString())
  }
  if (url.searchParams.has("sslcert")) {
    opts.dialectOptions.ssl.cert = fs.readFileSync(url.searchParams.get("sslcert").toString())
  }
  if (url.searchParams.has("sslkey")) {
    opts.dialectOptions.ssl.key = fs.readFileSync(url.searchParams.get("sslkey").toString())
  }
}
var sequelize = new Sequelize(opts);
var DataTypes = Sequelize.DataTypes;

if (!Sequelize.supportsCockroachDB) {
  throw new Error("CockroachDB dialect for Sequelize not installed");
}

module.exports.Customer = sequelize.define('customer', {
  name: DataTypes.STRING
}, {
  timestamps: false
});

module.exports.customerToJSON = function(customer) {
  return {
    id: parseInt(customer.id),
    name: customer.name
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
