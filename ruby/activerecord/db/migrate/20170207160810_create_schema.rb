class CreateSchema < ActiveRecord::Migration
  def change
    create_table :customers do |t|
      t.string :name

      t.timestamps null: false
    end

    create_table :products do |t|
      t.string :name
      t.decimal :price, precision: 18, scale: 2

      t.timestamps null: false
    end

    create_table :orders do |t|
      t.decimal :subtotal, precision: 18, scale: 2
      t.belongs_to :customer, index: true

      t.timestamps null: false
    end

    create_table :line_items do |t|
      t.belongs_to :order, index: true
      t.belongs_to :product, index: true

      t.timestamps null: false
    end
  end
end
