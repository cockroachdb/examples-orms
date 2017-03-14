var express = require('express');
var router = express.Router();
var models = require('../models');

// GET all orders from database.
router.get('/', function(req, res, next) {
  models.order.findAll().then(function(orders) {
    res.setHeader('Content-Type', 'application/json');
    var result = orders.map(function(order) { return order.jsonObject(); });
    res.send(result);
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

// POST a new order to the database.
router.post('/', function(req, res, next) {
  var o = {
    id: parseInt(req.body.id),
    subtotal: parseFloat(req.body.subtotal),
    customer_id: parseInt(req.body.customer.id)
  };
  models.order.create(o).then(function(order) {
    res.setHeader('Content-Type', 'application/json');
    res.send(order.jsonObject());
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

// GET one order using its id, from the database.
router.get('/:id', function(req, res, next) {
  var id = parseInt(req.params.id);
  models.order.findById(id).then(function(order) {
    res.setHeader('Content-Type', 'application/json');
    res.send(order.jsonObject());
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

module.exports = router;
