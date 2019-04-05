require_relative 'placeholders'

require 'json'

step "metrics for tenant :tenant should report :count :transaction_status transactions" do |tenant, count, status|
  metrics_file = "/reports/metrics.#{tenant}.json"

  eventually(timeout: 3) {
    expect(File.file?(metrics_file)).to be(true)
    metrics = File.open(metrics_file, 'rb') { |f| JSON.parse(f.read) }
    expect(metrics["#{status}Transactions"]).to eq(count)
  }
end

step "metrics file :path should have following keys:" do |path, keys|
  expected_keys = keys.split("\n").map(&:strip).reject { |x| x.empty? }

  eventually(timeout: 3) {
    expect(File.file?(path)).to be(true)
  }

  metrics_keys = File.open(path, 'rb') { |f| JSON.parse(f.read).keys }

  expect(metrics_keys).to match_array(expected_keys)
end
