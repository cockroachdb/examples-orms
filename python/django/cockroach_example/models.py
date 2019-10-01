from django.db import models

class Customers(models.Model):
    id = models.AutoField(primary_key=True)
    name = models.CharField(max_length=250)

class Products(models.Model):
    id = models.AutoField(primary_key=True)
    name = models.CharField(max_length=250)
    price = models.DecimalField(max_digits=18, decimal_places=2)

class Orders(models.Model):
    id = models.AutoField(primary_key=True)
    subtotal = models.DecimalField(max_digits=18, decimal_places=2)
    # TODO (rohany): is setting null the right thing here? The schema doesn't say what to do. 
    customer = models.ForeignKey(Customers, on_delete=models.SET_NULL, null=True)
    product = models.ManyToManyField(Products)

