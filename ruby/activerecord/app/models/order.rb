class Order < ActiveRecord::Base
  attr_accessible :subtotal

  belongs_to :customer
  has_many :line_items
  has_many :products, :through => :line_items
end
