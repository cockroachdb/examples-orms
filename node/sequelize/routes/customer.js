var express = require('express');
var router = express.Router();
var models = require('../models');

// GET all customers.
router.get('/', function(req, res, next) {
  models.customer.findAll().then(function(customers) {
    res.setHeader('Content-Type', 'application/json');
    var result = customers.map(function(customer) { return customer.jsonObject(); });
    res.send(result);
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

// POST a new customer.
router.post('/', function(req, res, next) {
  var c = {id: req.body.id, name: req.body.name};
  models.customer.create(c).then(function(customer) {
    res.setHeader('Content-Type', 'application/json');
    res.send(customer.jsonObject());
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

// GET one customer.
router.get('/:id', function(req, res, next) {
  var id = parseInt(req.params.id);
  models.customer.findById(id).then(function(customer) {
    res.setHeader('Content-Type', 'application/json');
    res.send(customer.jsonObject());
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

module.exports = router;
