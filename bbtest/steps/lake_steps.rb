
step "lake is empty" do ||
  LakeMock.reset()
end

step "lake recieves :data" do |data|
  LakeMock.send(data)
end

step "lake responds with :data" do |data|
  eventually() {
    ok = LakeMock.pulled_message?(data)
    expect(ok).to be(true), "message #{data} was not found in #{LakeMock.mailbox()}"
  }
  LakeMock.ack(data)
end

step "no other messages were relayed" do ||
  expect(LakeMock.mailbox()).to be_empty, "expected empty mailbox but got dangling messages: #{LakeMock.mailbox()}"
end
