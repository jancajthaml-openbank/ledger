#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import json
import time
import decimal
from openbank_testkit import Request


def create_transfer(context, tenant):
  uri = "https://127.0.0.1/transaction/{}".format(tenant)

  request = Request(method='POST', url=uri)
  request.add_header('Accept', 'application/json')
  request.add_header('Content-Type', 'application/json')

  request.data = context.http_request_body

  response = request.do()
  if response.status == 504:
    response = request.do()

  assert response.status in [200, 201, 417], str(response.status)

  if response.status in [200, 201]:
    context.last_transaction_id = response.read().decode('utf-8')
  else:
    context.last_transaction_id = None


@when('following transaction is created from tenant {tenant}')
def create_custom_transfer(context, tenant):
  context.http_request_body = context.text
  create_transfer(context, tenant)


@when('{amount} {currency} is transferred from {tenantFrom}/{accountFrom} to {tenantTo}/{accountTo}')
def create_simple_transfer(context, amount, currency, tenantFrom, accountFrom, tenantTo, accountTo):
  create_simple_transfer_with_id(context, amount, currency, tenantFrom, accountFrom, tenantTo, accountTo, None)


@when('{amount} {currency} is transferred with id {transaction} from {tenantFrom}/{accountFrom} to {tenantTo}/{accountTo}')
def create_simple_transfer_with_id(context, amount, currency, tenantFrom, accountFrom, tenantTo, accountTo, transaction):
  payload = {
    'transfers': [{
      'credit': {
        'name': accountTo,
        'tenant': tenantTo,
      },
      'debit': {
        'name': accountFrom,
        'tenant': tenantFrom,
      },
      'amount': amount,
      'currency': currency
    }]
  }

  if transaction:
    payload['id'] = transaction

  context.http_request_body = json.dumps(payload)
  create_transfer(context, tenantFrom)


@given('{tenant}/{account} balance should be {amount} {currency}')
@then('{tenant}/{account} balance should be {amount} {currency}')
def account_balance_should_be(context, tenant, account, amount, currency):
  snapshot = context.vault.get_account(tenant, account)

  assert snapshot, 'missing snapshot for {}/{}'.format(tenant, account)
  assert snapshot['currency'] == currency, 'currency mismatch expected {} actual {}'.format(currency, snapshot['currency'])
  assert snapshot['balance'] == decimal.Decimal(amount), 'balance mismatch expected {} actual {}'.format(decimal.Decimal(amount), snapshot['balance'])


@given('vault is empty')
def empty_vault(context):
  context.vault.reset()


@given('{activity} account {tenant}/{account} with currency {currency} exist')
@when('{activity} account {tenant}/{account} with currency {currency} exist')
def ensure_account(context, activity, currency, tenant, account):
  context.vault.create_account(tenant, account, "test", currency, activity == "active")


@when('I request HTTP {uri}')
def perform_http_request(context, uri):
  options = dict()
  if context.table:
    for row in context.table:
      options[row['key']] = row['value']

  request = Request(method=options['method'], url=uri)
  request.add_header('Accept', 'application/json')
  if context.text:
    request.add_header('Content-Type', 'application/json')
    request.data = context.text

  response = request.do()
  context.http_response = {
    'status': str(response.status),
    'body': response.read().decode('utf-8'),
    'content-type': response.content_type
  }


@then('HTTP response is')
def check_http_response(context):
  options = dict()
  if context.table:
    for row in context.table:
      options[row['key']] = row['value']

  assert context.http_response
  response = context.http_response
  del context.http_response

  if 'status' in options:
    assert response['status'] == options['status'], 'expected status {} actual {}'.format(options['status'], response['status'])

  if context.text:
    def diff(path, a, b):
      if type(a) == list:
        assert type(b) == list, 'types differ at {} expected: {} actual: {}'.format(path, list, type(b))
        for idx, item in enumerate(a):
          assert item in b, 'value {} was not found at {}[{}]'.format(item, path, idx)
          diff('{}[{}]'.format(path, idx), item, b[b.index(item)])
      elif type(b) == dict:
        assert type(b) == dict, 'types differ at {} expected: {} actual: {}'.format(path, dict, type(b))
        for k, v in a.items():
          assert k in b
          diff('{}.{}'.format(path, k), v, b[k])
      else:
        assert type(a) == type(b), 'types differ at {} expected: {} actual: {}'.format(path, type(a), type(b))
        assert a == b, 'values differ at {} expected: {} actual: {}'.format(path, a, b)

    stash = list()

    if response['body']:
      for line in response['body'].split('\n'):
        if response['content-type'].startswith('text/plain'):
          stash.append(line)
        else:
          stash.append(json.loads(line))

    try:
      expected = json.loads(context.text)
      if type(expected) == dict:
        stash = stash[0] if len(stash) else dict()
      diff('', expected, stash)
    except AssertionError as ex:
      raise AssertionError('{} with response {}'.format(ex, response['body']))
