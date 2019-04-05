
step "tenant :tenant receives :data" do |tenant, data|
  LakeMock.send("VaultUnit/#{tenant} Wall/bbtest #{data}")
end

step "tenant :tenant responds with :data" do |_, data|
  expected = LakeMock.parse_message(data)
  eventually() {
    ok = LakeMock.pulled_message?(expected)
    expect(ok).to be(true), "message #{expected} was not found in #{LakeMock.parsed_mailbox()}"
  }
  LakeMock.ack(expected)
end

step "no other messages were received" do ||
  expect(LakeMock.mailbox()).to be_empty, "expected empty mailbox but got dangling messages: #{LakeMock.mailbox()}"
end
