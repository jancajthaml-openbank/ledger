
step "package :package is installed" do |package|
  out = %x(apt-get -y install -f /etc/bbtest/packages/#{package})
  expect($?).to be_success, out

  expect(File.file?("/etc/init/ledger.conf")).to be(true)
end

step "package :package is uninstalled" do |package|
  out = %x(apt-get -y remove #{package})
  expect($?).to be_success, out

  expect(File.file?("/etc/init/ledger.conf")).to be(false)
end

step "systemctl contains following" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  eventually() {
    items.each { |item|
      units = %x(systemctl list-units --no-legend | awk '{ print $1 }')
      units = units.split("\n").map(&:strip).reject(&:empty?)
      subset = units.reject { |x| !x.include?(item) }
      expect(subset).not_to be_empty, "\"#{item}\" was not found in #{units}"
    }
  }
end

step "systemctl does not contains following" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  items.each { |item|
    units = %x(systemctl list-units --no-legend | awk '{ print $1 }')
    units = units.split("\n").map(&:strip).reject(&:empty?)
    subset = units.reject { |x| !x.include?(item) }
    expect(subset).to be_empty, "#{item} was found not found in #{units}"
  }
end

step ":operation unit :unit" do |operation, unit|
  eventually(timeout: 5) {
    %x(systemctl #{operation} #{unit} 2>&1)
  }

  unless $? == 0
    err = %x(systemctl status #{unit} 2>&1)
    raise "operation \"systemctl #{operation} #{unit}\" returned error: #{err}"
  end
end

step "unit :unit is running" do |unit|
  eventually() {
    out = %x(systemctl show -p SubState #{unit} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "unit :unit is not running" do |unit|
  eventually() {
    out = %x(systemctl show -p SubState #{unit} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).not_to eq("running")
  }
end
