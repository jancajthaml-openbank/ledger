require 'date'
require 'json'

module HTTPHelper

  class << self
    attr_accessor :request, :response
  end

  def self.prepare_request(options)
    cmd = ["curl"]
    cmd << ["--insecure"]
    cmd << ["--noproxy 'localhost,127.0.0.1'"]
    cmd << ["-ss"]
    cmd << ["-v"]
    cmd << ["-i"]
    unless options[:method].upcase == "GET"
      cmd << ["-X #{options[:method].upcase}"]
    end
    cmd << [options[:url]]
    unless options[:body].nil? or options[:method].upcase == "GET"
      cmd << ["-H \"Content-Type: application/json\""]
      cmd << ["-d \'#{JSON.parse(options[:body]).to_json}\'"]
    end
    cmd << ["2>&1 | cat"]

    self.request = cmd.join(" ")
  end

  def self.perform_request()
    raise "no request prepared" if self.request.nil?

    resp = %x(#{self.request})

    self.response = { :code => 0, :raw => resp }

    lines = resp.split("\n")
    lines.each_with_index { |line, idx|
      if line.start_with? "HTTP/"
        self.response[:code] = line[9..13].to_i
      elsif line.strip.empty?
        self.response[:body] = lines[(idx+1)...lines.length].join("\n")
        break
      end
    }

    raise "endpoint is unreachable\n#{self.response[:raw]}" if self.response[:code] === 0
    return
  end

end
