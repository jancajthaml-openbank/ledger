require_relative 'placeholders'

require 'json'

step "transaction of tenant :tenant should be" do |tenant, expectation|
  transaction = transaction(tenant, @transaction_id)
  expectation = JSON.parse(expectation)

  expect(transaction["id"]).to eq(expectation["id"]) unless expectation["id"].nil?

  expectation["transfers"].each { |e|
    found = false
    transaction["transfers"].each { |t|
      same = true
      same &&= t["id"] == e["id"] unless e["id"].nil?
      same &&= t["credit"] == e["credit"] unless e["credit"].nil?
      same &&= t["debit"] == e["debit"] unless e["debit"].nil?
      same &&= t["valueDate"] == e["valueDate"] unless e["valueDate"].nil?
      same &&= t["amount"] == e["amount"] unless e["amount"].nil?
      same &&= t["currency"] == e["currency"] unless e["currency"].nil?

      if same
        found = true
        break
      end
    }
    raise "#{e} not found in #{transaction}" unless found
  }
end

step "transaction of tenant :tenant should not exist" do |tenant|
  return if @transaction_id.nil?
  expect(transaction_data(tenant, @transaction_id)).to be_nil, "transaction found #{tenant} #{@transaction_id}"
end

step "transaction of tenant :tenant should exist" do |tenant|
  expect(@transaction_id).not_to be_nil
  expect(transaction_data(tenant, @transaction_id)).not_to be_nil, "transaction not found #{tenant} #{@transaction_id}"
end
