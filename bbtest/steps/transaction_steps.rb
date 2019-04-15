require_relative 'placeholders'

require 'json'

step ":amount :currency is transferred from :account to :account" do |amount, currency, from, to|
  (fromTenant, fromAccount) = from.split('/')
  (toTenant, toAccount) = to.split('/')

  payload = {
    transfers: [{
      credit: {
        name: toAccount,
        tenant: toTenant,
      },
      debit: {
        name: fromAccount,
        tenant: fromTenant,
      },
      amount: amount,
      currency: currency
    }]
  }.to_json

  send "following transaction is created from tenant :tenant", fromTenant, payload
end

step "following transaction is created from tenant :tenant" do |tenant, data = nil|
  uri = "https://127.0.0.1/transaction/#{tenant}"

  send "I request curl :http_method :url", "POST", uri, data

  @resp = { :code => 0 }

  eventually(timeout: 60, backoff: 2) {
    resp = %x(#{@http_req})
    @resp[:code] = resp[resp.length-3...resp.length].to_i

    if @resp[:code] === 0
      raise "endpoint #{@http_req} is unreachable"
    end

    @resp[:body] = resp[0...resp.length-3] unless resp.nil?

    case @resp[:code]
      when 200, 201
        @transaction_id = JSON.parse(@resp[:body])["id"]
      else
        @transaction_id = nil
    end
  }
end

step ":id :id :side side is forwarded to :account from tenant :tenant" do |transaction, transfer, side, account, tenant|
  (tenant, account) = account.split('/')

  payload = {
    side: side,
    target: {
      tenant: tenant,
      name: account
    }
  }.to_json

  uri = "https://127.0.0.1/transaction/#{tenant}/#{transaction}/#{transfer}"

  send "I request curl :http_method :url", "PATCH", uri, payload

  @resp = { :code => 0 }

  eventually(timeout: 60, backoff: 2) {
    resp = %x(#{@http_req})
    @resp[:code] = resp[resp.length-3...resp.length].to_i

    if @resp[:code] === 0
      raise "endpoint #{@http_req} is unreachable"
    end

    @resp[:body] = resp[0...resp.length-3] unless resp.nil?
  }
end

step "request should succeed" do ||
  expect(@resp[:code]).to eq(200), "#{@http_req} -> #{@resp[:code]} #{@resp[:body]}"
end

step "request should fail" do ||
  expect(@resp[:code]).to_not eq(200), "#{@http_req} -> #{@resp[:code]} #{@resp[:body]}"
end
