require 'date'

module JournalHelper

  def self.transaction(tenant, id)
    return nil if id.nil?
    path = "/data/t_#{tenant}/transaction/#{id}"
    return nil unless File.file?(path)

    File.open(path, 'rb') { |f|
      lines = f.read.split("\n").map(&:strip)
      transfers = lines[1..-1]
      transfers = [] if transfers.nil?

      {
        "id" => id,
        "state" => lines[0],
        "transfers" => transfers.map { |line|
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
