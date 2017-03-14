'use strict';
module.exports = function(sequelize, DataTypes) {
  var Product = sequelize.define('product', {
    name: DataTypes.STRING,
    price: DataTypes.DECIMAL(18, 2)
  }, {
    instanceMethods: {
      // jsonObject returns an object that JSON.stringify can trivially
      // convert into the JSON response the test driver expects.
      jsonObject: function() {
        return {
          id: parseInt(this.id),
          name: this.name,
          price: this.price,
        };
      }
    },
    timestamps: false
  });
  return Product;
};
