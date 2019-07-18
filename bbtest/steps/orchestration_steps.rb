require_relative 'placeholders'

step "ledger is restarted" do ||
  ids = %x(systemctl -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("ledger-unit@")
  }.map { |x| x.chomp(".service") }

  ids << "ledger-rest"

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "ledger is running" do ||
  ids = %x(systemctl -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("ledger-unit@")
  }.map { |x| x.chomp(".service") }

  ids << "ledger-rest"

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "tenant :tenant is offboarded" do |tenant|
  eventually() {
    %x(journalctl -o short-precise -u ledger-unit@#{tenant}.service --no-pager > /tmp/reports/ledger-unit@#{tenant}.log 2>&1)
    %x(systemctl stop ledger-unit@#{tenant} 2>&1)
    %x(systemctl disable ledger-unit@#{tenant} 2>&1)
    %x(journalctl -o short-precise -u ledger-unit@#{tenant}.service --no-pager > /tmp/reports/ledger-unit@#{tenant}.log 2>&1)
  }
end

step "tenant :tenant is onbdoarded" do |tenant|
  config = Array[UnitHelper.default_config.map { |k,v| "LEDGER_#{k}=#{v}" }]
  config = config.join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{config}' > /etc/init/ledger.conf)

  %x(systemctl enable ledger-unit@#{tenant} 2>&1)
  %x(systemctl start ledger-unit@#{tenant} 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState ledger-unit@#{tenant} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "ledger is reconfigured with" do |configuration = ""|
  params = Hash[configuration.split("\n").map(&:strip).reject(&:empty?).map { |el| el.split '=' }]
  config = Array[UnitHelper.default_config.merge(params).map { |k,v| "LEDGER_#{k}=#{v}" }]
  config = config.join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{config}' > /etc/init/ledger.conf)

  ids = %x(systemctl list-units | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("ledger-")
  }.map { |x| x.chomp(".service") }

  expect(ids).not_to be_empty

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end
