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
  HTTPHelper.prepare_request({
    :method => "POST",
    :url => "https://127.0.0.1/transaction/#{tenant}",
    :body => data
  })

  eventually(timeout: 60, backoff: 2) {
    HTTPHelper.perform_request()

    case HTTPHelper.response[:code]
      when 200, 201
        @transaction_id = JSON.parse(HTTPHelper.response[:body])["id"]
      else
        @transaction_id = nil
    end
  }
end

step ":id :id :side side is forwarded to :account from tenant :tenant" do |transaction, transfer, side, account, tenant|
  (tenant, account) = account.split('/')

  HTTPHelper.prepare_request({
    :method => "PATCH",
    :url => "https://127.0.0.1/transaction/#{tenant}/#{transaction}/#{transfer}",
    :body => {
      side: side,
      target: {
        tenant: tenant,
        name: account
      }
    }.to_json
  })

  eventually(timeout: 60, backoff: 2) {
    HTTPHelper.perform_request()
  }
end

step "request should succeed" do ||
  expect(HTTPHelper.response[:code]).to eq(200), "expected 200 got\n#{HTTPHelper.response[:raw]}"
end

step "request should fail" do ||
  expect(HTTPHelper.response[:code]).to_not eq(200), "expected non 200 got\n#{HTTPHelper.response[:raw]}"
end
