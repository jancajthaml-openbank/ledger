require_relative 'placeholders'

require 'bigdecimal'

step "vault is empty" do ||
  VaultHelper.reset()
end

step ":activity account :account with currency :currency exist" do |activity, account, currency|
  (tenant, account) = account.split('/')

  VaultHelper.create_account(tenant, account, currency, activity)
end

step ":account balance should be :amount :currency" do |account, amount, currency|
  (tenant, account) = account.split('/')

  snapshot = VaultHelper.get_acount(tenant, account)

  expect(snapshot[:currency]).to eq(currency)
  expect(BigDecimal.new(snapshot[:balance]).to_s('F')).to eq(BigDecimal.new(amount).to_s('F'))
end
