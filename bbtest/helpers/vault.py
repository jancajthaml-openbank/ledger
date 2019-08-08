#!/usr/bin/env python
# -*- coding: utf-8 -*-
from decimal import Decimal


class VaultHelper(object):

  def __init__(self, context):
    self.tenants = dict()
    self.context = context

  def reset(self):
    self.tenants = dict()

  def get_account(self, tenant, account):
    if not self.account_exist(tenant, account):
      return {}
    return self.tenants[tenant][account]

  def account_exist(self, tenant, account):
    return tenant in self.tenants and account in self.tenants[tenant]

  def create_account(self, tenant, account, format, currency, is_balance_check):
    if self.account_exist(tenant, account):
      return False
    if not tenant in self.tenants:
      self.tenants[tenant] = dict()
    self.tenants[tenant][account] = {
      'format': format,
      'currency': currency,
      'is_balance_check': is_balance_check,
      'balance': Decimal('0'),
      'blocking': Decimal('0'),
      'promised': dict()
    }
    return True

  def __process_promise_order(self, tenant, account, transaction, amount, currency):
    if not self.account_exist(tenant, account):
      return 'EE'
    if transaction in self.tenants[tenant][account]['promised']:
      return 'P1'
    if currency != self.tenants[tenant][account]['currency']:
      return 'P2 CURRENCY_MISMATCH'

    want = Decimal(amount)

    if self.tenants[tenant][account]['is_balance_check'] and (want + self.tenants[tenant][account]['balance']).is_signed():
      return 'P2 INSUFFICIENT_FUNDS'

    self.tenants[tenant][account]['promised'][transaction] = want
    self.tenants[tenant][account]['balance'] += want
    self.tenants[tenant][account]['blocking'] -= want

    return 'P1'

  def __process_commit_order(self, tenant, account, transaction):
    if not self.account_exist(tenant, account):
      return 'EE'
    if not transaction in self.tenants[tenant][account]['promised']:
      return 'C1'

    promised = self.tenants[tenant][account]['promised'][transaction]

    self.tenants[tenant][account]['blocking'] += promised
    del self.tenants[tenant][account]['promised'][transaction]

    return 'C1'

  def __process_rollback_order(self, tenant, account, transaction):
    if not self.account_exist(tenant, account):
      return 'R1'
    if not transaction in self.tenants[tenant][account]['promised']:
      return 'R1'

    promised = self.tenants[tenant][account]['promised'][transaction]

    self.tenants[tenant][account]['balance'] -= promised
    self.tenants[tenant][account]['blocking'] += promised
    del self.tenants[tenant][account]['promised'][transaction]

    return 'R1'

  def process_account_event(self, tenant, account, kind, transaction, amount, currency):
    if kind == 'NP':
      return self.__process_promise_order(tenant, account, transaction, amount, currency)
    elif kind == 'NC':
      return self.__process_commit_order(tenant, account, transaction)
    elif kind == 'NR':
      return self.__process_rollback_order(tenant, account, transaction)
    else:
      return 'EE'
