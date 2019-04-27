require 'ffi-rzmq'
require 'thread'
require 'bigdecimal'
require 'timeout'

module LakeMock

  def self.start
    raise "cannot start when shutting down" if self.poisonPill
    self.poisonPill = false
    self.mutex = Mutex.new
    self.recv_backlog = []

    begin
      ctx = ZMQ::Context.new
      pull_channel = ctx.socket(ZMQ::PULL)
      raise "unable to bind PULL" unless pull_channel.bind("tcp://*:5562") >= 0
      pub_channel = ctx.socket(ZMQ::PUB)
      raise "unable to bind PUB" unless pub_channel.bind("tcp://*:5561") >= 0
    rescue ContextError => _
      raise "Failed to allocate context or socket!"
    end

    self.ctx = ctx
    self.pull_channel = pull_channel
    self.pub_channel = pub_channel

    self.pull_daemon = Thread.new {
      loop do
        data = ""
        begin
          Timeout.timeout(1) do
            self.pull_channel.recv_string(data, 0)
          end
        rescue Timeout::Error => _
          break if self.poisonPill or self.pull_channel.nil?
          next
        end
        next if data.empty?

        if data.end_with?("]")
          self.pub_channel.send_string(data)
          self.pub_channel.send_string(data)
          next
        end

        self.mutex.synchronize do
          self.recv_backlog << data
        end

        unless data.start_with?("VaultUnit/")
          self.send(data)
          next
        end

        self.process_next_message(data)
      end
    }
  end

  def self.mailbox()
    return self.recv_backlog
  end

  def self.pulled_message?(expected)
    copy = self.recv_backlog.dup
    copy.each { |item|
      return true if item == expected
    }
    return false
  end

  def self.ack(data)
    self.mutex.synchronize do
      self.recv_backlog.reject! { |v| v == data }
    end
  end

  def self.reset()
    self.mutex.synchronize do
      self.recv_backlog = []
    end
  end

  def self.process_next_message(data)
    if groups = data.match(/^VaultUnit\/([^\s]{1,100}) LedgerUnit\/([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) (-?\d{1,100}\.\d{1,100}|-?\d{1,100}) ([A-Z]{3})$/i)
      tenant, sender, account, req_id, kind, transaction, amount, currency = groups.captures
      reply_event = VaultHelper.process_account_event(tenant, account, kind, transaction, amount, currency)

      self.reply(sender, tenant, account, req_id, reply_event)
    else
      puts "!!!!!! unknown message received !!!!!! #{data}"
    end
  end

  def self.reply(target, tenant, sender, replyTo, msg = "EE")
    self.send("LedgerUnit/#{target} VaultUnit/#{tenant} #{replyTo} #{sender} #{msg}")
  end

  def self.send(data)
    self.pub_channel.send_string(data) unless self.pub_channel.nil?
  end

  def self.stop
    self.poisonPill = true
    begin
      self.pull_daemon.join() unless self.pull_daemon.nil?
      self.pub_channel.close() unless self.pub_channel.nil?
      self.pull_channel.close() unless self.pull_channel.nil?
      self.ctx.terminate() unless self.ctx.nil?
    rescue
    ensure
      self.pull_daemon = nil
      self.ctx = nil
      self.pull_channel = nil
      self.pub_channel = nil
    end
    self.poisonPill = false
  end

  class << self
    attr_accessor :ctx,
                  :pull_channel,
                  :pub_channel,
                  :pull_daemon,
                  :poisonPill,
                  :mutex,
                  :recv_backlog
  end

end
