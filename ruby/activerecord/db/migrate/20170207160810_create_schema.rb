class CreateSchema < ActiveRecord::Migration[5.1]
  def change
    create_table :customers do |t|
      t.string :name, null: false
    end

    create_table :products do |t|
      t.string :name, null: false
      t.decimal :price, precision: 18, scale: 2, null: false
    end

    create_table :orders do |t|
      t.decimal :subtotal, precision: 18, scale: 2, null: false, default: 0
      t.belongs_to :customer, index: true, null: false
    end

    add_foreign_key :orders, :customers

    create_table :order_products, id: false do |t|
      t.belongs_to :order, index: true, null: false
      t.belongs_to :product, index: true, null: false
    end

    add_foreign_key :order_products, :orders
    add_foreign_key :order_products, :products
  end
end
