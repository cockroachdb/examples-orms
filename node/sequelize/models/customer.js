'use strict';
module.exports = function(sequelize, DataTypes) {
  var Customer = sequelize.define('customer', {
    name: DataTypes.STRING
  }, {
    instanceMethods: {
      // jsonObject returns an object that JSON.stringify can trivially
      // convert into the JSON response the test driver expects.
      jsonObject: function() {
        return {
          id: parseInt(this.id),
          name: this.name
        };
      }
    },
    timestamps: false
  });
  return Customer;
};
