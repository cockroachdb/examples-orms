# encoding: UTF-8
# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# Note that this schema.rb definition is the authoritative source for your
# database schema. If you need to create the application database on another
# system, you should be using db:schema:load, not running all the migrations
# from scratch. The latter is a flawed and unsustainable approach (the more migrations
# you'll amass, the slower it'll run and the greater likelihood for issues).
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 20170207160810) do

  create_table "customers", id: :bigserial, force: :cascade do |t|
    t.string "name", null: false
  end

# Could not dump table "order_products" because of following ArgumentError
#   struct size differs

# Could not dump table "orders" because of following ArgumentError
#   struct size differs

  create_table "products", id: :bigserial, force: :cascade do |t|
    t.string  "name",                           null: false
    t.decimal "price", precision: 18, scale: 2, null: false
  end

  add_foreign_key "order_products", "orders"
  add_foreign_key "order_products", "products"
  add_foreign_key "orders", "customers"
end
