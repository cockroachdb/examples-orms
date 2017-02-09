<<<<<<< HEAD
class Product < ApplicationRecord
  has_and_belongs_to_many :orders, join_table: 'order_products'
||||||| parent of 4657ab0... Models inherit from application_record
class Product < ActiveRecord::Base
  has_many :line_items
  has_many :orders, :through => :line_items
=======
class Product < ApplicationRecord
  has_many :line_items
  has_many :orders, :through => :line_items
>>>>>>> 4657ab0... Models inherit from application_record
end
