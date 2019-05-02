require_relative 'eventually_helper'

require 'fileutils'
require 'timeout'
require 'thread'

Thread.abort_on_exception = true

Encoding.default_external = Encoding::UTF_8
Encoding.default_internal = Encoding::UTF_8

class UnitHelper

  attr_reader :units

  def download()
    raise "no version specified" unless ENV.has_key?('UNIT_VERSION')

    version = ENV['UNIT_VERSION']

    FileUtils.mkdir_p "/opt/artifacts"
    %x(rm -rf /opt/artifacts/*)

    FileUtils.mkdir_p "/etc/bbtest/packages"
    %x(rm -rf /etc/bbtest/packages/*)

    %x(docker run --name temp-container-ledger openbank/ledger:#{version} /bin/true)
    %x(docker cp temp-container-ledger:/opt/artifacts/. /opt/artifacts)
    %x(docker rm temp-container-ledger)

    Dir.glob('/opt/artifacts/ledger_*_amd64.deb').each { |f|
      puts "#{f}"
      FileUtils.mv(f, '/etc/bbtest/packages/ledger.deb')
    }

    raise "no package to install" unless File.file?('/etc/bbtest/packages/ledger.deb')
  end

  def cleanup()
    %x(systemctl -t service --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("ledger") }
      .map { |x| x.chomp(".service") }
      .each { |unit|
        if unit.start_with?("ledger-unit@")
          %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
          %x(systemctl stop #{unit} 2>&1)
          %x(systemctl disable #{unit} 2>&1)
          %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        else
          %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        end
      }
  end

  def teardown()
    %x(systemctl -t service --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("ledger") }
      .map { |x| x.chomp(".service") }
      .each { |unit|
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        %x(systemctl stop #{unit} 2>&1)
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)

        if unit.include?("@")
          metrics_file = "/opt/#{unit[/[^@]+/]}/metrics/metrics.#{unit[/([^@]+)$/]}.json"
        else
          metrics_file = "/opt/#{unit}/metrics/metrics.json"
        end

        File.open(metrics_file, 'rb') { |fr|
          File.open("/reports/metrics/#{unit.gsub('@','_')}.json", 'w') { |fw|
            fw.write(fr.read)
          }
        } if File.file?(metrics_file)
      }
  end

end
