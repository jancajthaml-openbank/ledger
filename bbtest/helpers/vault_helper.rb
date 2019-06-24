require 'bigdecimal'

module VaultHelper

  class << self
    attr_accessor :tenants
  end

  self.tenants = Hash.new()

  def self.reset()
    self.tenants = Hash.new()
  end

  def self.get_acount(tenant, id)
    return {} unless self.tenants.has_key?(tenant)
    return {} unless self.tenants[tenant].has_key?(id)
    return self.tenants[tenant][id]
  end

  def self.account_exist?(tenant, id)
    return self.tenants.has_key?(tenant) && self.tenants[tenant].has_key?(id)
  end

  def self.create_account(tenant, id, format, currency, is_balance_check)
    return false if self.tenants.has_key?(tenant) && self.tenants[tenant].has_key?(id)

    self.tenants[tenant] = Hash.new() unless self.tenants.has_key?(tenant)
    self.tenants[tenant][id] = {
      :format => format,
      :currency => currency,
      :is_balance_check => is_balance_check,
      :balance => BigDecimal.new("0"),
      :blocking => BigDecimal.new("0"),
      :promised => {}
    }
    return true
  end

  def self.process_promise_order(tenant, id, transaction, amount, currency)
    return "EE" unless (self.tenants.has_key?(tenant) && self.tenants[tenant].has_key?(id))
    return "P1" if self.tenants[tenant][id][:promised].has_key?(transaction)
    return "P2 CURRENCY_MISMATCH" unless currency == self.tenants[tenant][id][:currency]

    want = BigDecimal.new(amount)

    return "P2 INSUFFICIENT_FUNDS" unless (!self.tenants[tenant][id][:is_balance_check] or (want + self.tenants[tenant][id][:balance]).sign() >= 0)

    self.tenants[tenant][id][:promised][transaction] = want
    self.tenants[tenant][id][:balance] = self.tenants[tenant][id][:balance] + want
    self.tenants[tenant][id][:blocking] = self.tenants[tenant][id][:blocking] - want

    return "P1"
  end

  def self.process_commit_order(tenant, id, transaction)
    return "EE" unless (self.tenants.has_key?(tenant) && self.tenants[tenant].has_key?(id))
    return "C1" unless self.tenants[tenant][id][:promised].has_key?(transaction)

    promised = self.tenants[tenant][id][:promised][transaction]

    self.tenants[tenant][id][:blocking] = self.tenants[tenant][id][:blocking] + promised
    self.tenants[tenant][id][:promised].tap { |hs| hs.delete(transaction) }

    return "C1"
  end

  def self.process_rollback_order(tenant, id, transaction)
    return "R1" unless (self.tenants.has_key?(tenant) && self.tenants[tenant].has_key?(id))
    return "R1" unless self.tenants[tenant][id][:promised].has_key?(transaction)

    promised = self.tenants[tenant][id][:promised][transaction]

    self.tenants[tenant][id][:balance] = self.tenants[tenant][id][:balance] - promised
    self.tenants[tenant][id][:blocking] = self.tenants[tenant][id][:blocking] + promised
    self.tenants[tenant][id][:promised].tap { |hs| hs.delete(transaction) }

    return "R1"
  end

  def self.process_account_event(tenant, id, kind, transaction, amount, currency)
    case kind
    when "NP" ;
      return self.process_promise_order(tenant, id, transaction, amount, currency)
    when "NC" ;
      return self.process_commit_order(tenant, id, transaction)
    when "NR" ;
      return self.process_rollback_order(tenant, id, transaction)
    else ; return "EE"
    end
  end

end
