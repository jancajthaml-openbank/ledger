
step "journalctl of :unit contains following" do |unit, expected|
  expected_lines = expected.split("\n").map(&:strip).reject(&:empty?)

  eventually() {
    actual = %x(journalctl -o short-precise -u #{unit} --no-pager 2>&1)
    expect($?).to be_success

    actual_lines_merged = actual.split("\n").map(&:strip).reject(&:empty?)
    actual_lines = []
    idx = actual_lines_merged.length - 1

    loop do
      break if idx < 0 or actual_lines_merged[idx].include?("INFO >>> Start <<<")
      actual_lines << actual_lines_merged[idx]
      idx -= 1
    end

    actual_lines = actual_lines.reverse

    expected_lines.each { |line|
      found = false
      actual_lines.each { |l|
        next unless l.include? line
        found = true
        break
      }
      raise "#{line} was not found in logs:\nlast:\n#{actual_lines.join("\n")}\nfull:\n#{actual_lines_merged.join("\n")}" unless found
    }
  }
end
