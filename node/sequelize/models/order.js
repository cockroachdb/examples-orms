'use strict';
module.exports = function(sequelize, DataTypes) {
  var Order = sequelize.define('order', {
    customer_id: DataTypes.INTEGER,
    subtotal: DataTypes.DECIMAL(18, 2)
  }, {
    instanceMethods: {
      // jsonObject returns an object that JSON.stringify can trivially
      // convert into the JSON response the test driver expects.
      jsonObject: function() {
        return{
          id: parseInt(this.id),
          subtotal: this.subtotal,
          customer: {
            id: this.customer_id
          }
        };
      }
    },
    timestamps: false
  });
  return Order;
};
