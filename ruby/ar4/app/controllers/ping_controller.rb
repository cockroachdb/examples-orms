class PingController < ApplicationController
  def ping
    render plain: "ruby/ar4"
  end
end
