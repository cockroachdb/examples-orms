class Product < ApplicationRecord
  has_and_belongs_to_many :orders, join_table: 'order_products'
end
