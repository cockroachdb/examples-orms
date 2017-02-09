class ProductsController < ApplicationController
  def index
    render json: Product.all
  end

  def show
    render json: Product.find(params[:id])
  end

  def create
    redirect_to Product.create!(product_params)
  end

  def update
    Product.find(params[:id]).update(product_params)
  end

  def destroy
    Product.find(params[:id]).destroy
    render plain: "ok"
  end

  def add_product_to_order
    o = Order.find(params[:order_id])
    p = Product.find(params[:product_id])
    o.products<<p
    o.subtotal += p.price
    o.save
    render plain: "ok"
  end

  private
    def product_params
      params.require(:product).permit(:name, :price)
    end
end
