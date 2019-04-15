require 'json-diff'
require 'deepsort'
require 'json'

step "I request curl :http_method :url" do |http_method, url, body = nil|
  cmd = ["curl --insecure"]
  cmd << ["-X #{http_method.upcase}"] unless http_method.upcase == "GET"
  cmd << ["#{url} -sw \"%{http_code}\""]
  cmd << ["-d \'#{JSON.parse(body).to_json}\'"] unless body.nil? or http_method.upcase == "GET"

  @http_req = cmd.join(" ")
end

step "curl responds with :http_status" do |http_status, body = nil|
  raise if @http_req.nil?

  @resp = { :code => 0 }

  eventually(timeout: 60, backoff: 2) {
    resp = %x(#{@http_req})
    @resp[:code] = resp[resp.length-3...resp.length].to_i

    if @resp[:code] === 0
      raise "endpoint #{@http_req} is unreachable"
    end

    http_status = [http_status] unless http_status.kind_of?(Array)
    expect(http_status).to include(@resp[:code])

    @resp[:body] = resp[0...resp.length-3] unless resp.nil?
  }

  return if body.nil?

  expectation = JSON.parse(body)
  expectation.deep_sort!

  begin
    resp_body = JSON.parse(@resp[:body])
    resp_body.deep_sort!

    diff = JsonDiff.diff(resp_body, expectation).select { |item| item["op"] == "add" }.map { |item| item["value"] or item }
    return if diff == []

    raise "expectation failure:\ngot:\n#{JSON.pretty_generate(resp_body)}\nexpected:\n#{JSON.pretty_generate(expectation)}\ndiff:#{JSON.pretty_generate(diff)}"

  rescue JSON::ParserError
    raise "invalid response got \"#{@resp[:body].strip}\", expected \"#{expectation.to_json}\""
  end
end

step "curl does not responds with :http_status" do |http_status|
  raise if @http_req.nil?

  @resp = { :code => 0 }

  eventually(timeout: 10, backoff: 1) {
    resp = %x(#{@http_req})
    @resp[:code] = resp[resp.length-3...resp.length].to_i
    raise "endpoint unreachable" if @resp[:code] === 0
    @resp[:body] = resp[0...resp.length-3] unless resp.nil?
  }

  expect(@resp[:code]).not_to eq(http_status)
end
