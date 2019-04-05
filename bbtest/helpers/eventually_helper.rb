require 'timeout'

module EventuallyHelper

  def self.eventually(timeout: 10, backoff: nil, &_blk)
    return unless block_given?
    last_err = nil
    begin
      Timeout.timeout(timeout) do
        begin
          yield
        rescue Exception => e
          last_err = e
          sleep backoff if backoff
          retry
        end
      end
    rescue Timeout::Error
      raise last_err if last_err
      raise "function timeout after #{timeout} seconds"
    end
  end

  def eventually(*args, &blk)
    EventuallyHelper.eventually(*args, &blk)
  end

end
