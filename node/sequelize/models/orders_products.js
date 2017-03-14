'use strict';

module.exports = function(sequelize, DataTypes) {
  var OrderProduct = sequelize.define('order_products', {
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
  return OrderProduct;
};
