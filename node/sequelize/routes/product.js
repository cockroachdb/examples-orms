var express = require('express');
var router = express.Router();
var models = require('../models');

// GET all products from database.
router.get('/', function(req, res, next) {
  models.product.findAll().then(function(products) {
    res.setHeader('Content-Type', 'application/json');
    var result = products.map(function(product) { return product.jsonObject(); });
    res.send(result);
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

// POST a new product to the database.
router.post('/', function(req, res, next) {
  var id = req.body.id || null;
  var p = {
    id: id,
    name: req.body.name,
    price: parseFloat(req.body.price)
  };
  models.product.create(p).then(function(product) {
    res.setHeader('Content-Type', 'application/json');
    res.send(product.jsonObject());
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

// GET one product using its id, from the database.
router.get('/:id', function(req, res, next) {
  var id = parseInt(req.params.id);
  models.product.findById(id).then(function(product) {
    res.setHeader('Content-Type', 'application/json');
    res.send(product.jsonObject());
  }).catch(function(e) {
    res.status(500).send(e);
  });
});

module.exports = router;
