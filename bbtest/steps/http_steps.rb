require 'json-diff'
require 'deepsort'
require 'json'

step "I request curl :http_method :url" do |http_method, url, body = nil|
  HTTPHelper.prepare_request({
    :method => http_method,
    :url => url,
    :body => body
  })
end

step "curl responds with :http_status" do |http_status, body = nil|
  eventually(timeout: 30, backoff: 2) {
    HTTPHelper.perform_request()
    http_status = [http_status] unless http_status.kind_of?(Array)
    expect(http_status).to include(HTTPHelper.response[:code])
  }

  return if body.nil?

  expectation = JSON.parse(body)
  expectation.deep_sort!

  begin
    resp_body = JSON.parse(HTTPHelper.response[:body])
    resp_body.deep_sort!

    diff = JsonDiff.diff(resp_body, expectation).select { |item| item["op"] == "add" || item["op"] == "replace" }.map { |item| item["value"] or item }
    return if diff == []

    raise "expectation failure:\ngot:\n#{JSON.pretty_generate(resp_body)}\nexpected:\n#{JSON.pretty_generate(expectation)}\ndiff:#{JSON.pretty_generate(diff)}"

  rescue JSON::ParserError
    raise "invalid response got \"#{HTTPHelper.response[:body].strip}\", expected \"#{expectation.to_json}\""
  end
end

step "curl does not responds with :http_status" do |http_status|
  eventually(timeout: 30, backoff: 2) {
    HTTPHelper.perform_request()
  }

  http_status = [http_status] unless http_status.kind_of?(Array)
  expect(http_status).not_to include(HTTPHelper.response[:code])
end
