class Customer < ApplicationRecord
  validates :name, presence: true
end
