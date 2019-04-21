require 'date'

module JournalHelper

  def self.transaction(tenant, id)
    transaction = self.transaction_data(tenant, id)
    state = self.transaction_state(tenant, id)

    return nil if (transaction.nil? or state.nil?)

    transaction[:state] = state
    transaction
  end

  def self.transaction_state(tenant, id)
    return nil if id.nil?
    path = "/data/t_#{tenant}/transaction_state/#{id}"
    raise "transaction state for #{id} not found" unless File.file?(path)
    File.open(path, 'rb') { |f| f.read }
  end

  def self.transaction_data(tenant, id)
    return nil if id.nil?
    path = "/data/t_#{tenant}/transaction/#{id}"
    return nil unless File.file?(path)

    File.open(path, 'rb') { |f|
      lines = f.read.split("\n").map(&:strip)

      {
        "id" => lines[0],
        "transfers" => lines[1..-1].map { |line|
          data = line.split(" ").map(&:strip)

          {
            "id" => data[0],
            "credit" => data[1],
            "debit" => data[2],
            "amount" => data[4],
            "currency" => data[5],
            "valueDate" => Time.at(data[3].to_i).to_datetime.strftime("%Y-%m-%dT%H:%M:%SZ")
          }
        }
      }
    }
  end

end
