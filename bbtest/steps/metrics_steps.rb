require_relative 'placeholders'

require 'deepsort'
require 'json'

step "metrics file :path has permissions :permissions" do |path, permissions|
  expect(File.file?(path)).to be(true)

  actual = File.stat(path).mode.to_s(8).split('')[-4..-1].join
  expect(actual).to eq(permissions)
end

step "metrics file :path should have following keys:" do |path, keys|
  expected_keys = keys.split("\n").map(&:strip).reject { |x| x.empty? }

  eventually(timeout: 3) {
    expect(File.file?(path)).to be(true)
  }

  metrics_keys = File.open(path, 'rb') { |f| JSON.parse(f.read).keys }

  expect(metrics_keys).to match_array(expected_keys)
end

step "metrics file :path reports:" do |path, data|
  eventually(timeout: 3, backoff: 0.5) {
    expect(File.file?(path)).to be(true)
  }

  expected_data = data.split("\n").map(&:strip)
    .reject { |x| x.empty? }
    .map { |l| l.chomp.split(' ', 2) }
    .map { |k,v| [k,v.to_i] }

  expected_data = Hash[expected_data]
  expected_data.deep_sort!

  eventually(timeout: 3, backoff: 1) {
    metrics_data = File.open(path, 'rb') { |f| JSON.parse(f.read) }
    metrics_data.deep_sort!

    expect(metrics_data).to eq(expected_data)
  }
end
