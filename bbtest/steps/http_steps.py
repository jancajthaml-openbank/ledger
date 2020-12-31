#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import ssl
import urllib.request
import socket
import http
import json
import time
import decimal
import os


def create_transfer(context, tenant):
  uri = "https://127.0.0.1/transaction/{}".format(tenant)

  ctx = ssl.create_default_context()
  ctx.check_hostname = False
  ctx.verify_mode = ssl.CERT_NONE

  request = urllib.request.Request(method='POST', url=uri)
  request.add_header('Accept', 'application/json')
  request.add_header('Content-Type', 'application/json')

  request.data = context.http_request_body.encode('utf-8')

  try:
    response = urllib.request.urlopen(request, timeout=10, context=ctx)
    assert response.code in [200, 201], response.code
    response = response.read().decode('utf-8')
    context.last_transaction_id = response or None
  except (http.client.RemoteDisconnected, socket.timeout):
    raise AssertionError('timeout')
  except urllib.error.HTTPError as err:
    assert err.code == 417, 'missing transaction id with {}'.format(ex)
    context.last_transaction_id = None


@then('transaction of tenant {tenant} should not exist')
def transaction_should_not_exist(context, tenant):
  if context.last_transaction_id:
    path = '/data/t_{}/transaction/{}'.format(tenant, context.last_transaction_id)
    assert os.path.isfile(path) is False, "{} exists but should not".format(path)


@then('transaction of tenant {tenant} should exist')
def transaction_should_exist(context, tenant):
  assert context.last_transaction_id, 'missing last_transaction_id from context'
  path = '/data/t_{}/transaction/{}'.format(tenant, context.last_transaction_id)
  assert os.path.isfile(path) is True, "{} does not exists but should".format(path)


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

  ctx = ssl.create_default_context()
  ctx.check_hostname = False
  ctx.verify_mode = ssl.CERT_NONE

  request = urllib.request.Request(method=options['method'], url=uri)
  request.add_header('Accept', 'application/json')
  if context.text:
    request.add_header('Content-Type', 'application/json')
    request.data = context.text.encode('utf-8')

  context.http_response = dict()

  try:
    response = urllib.request.urlopen(request, timeout=10, context=ctx)
    context.http_response['status'] = str(response.status)
    context.http_response['body'] = response.read().decode('utf-8')
    context.http_response['content-type'] = response.info().get_content_type()
  except (http.client.RemoteDisconnected, socket.timeout):
    context.http_response['status'] = '504'
    context.http_response['body'] = ""
    context.http_response['content-type'] = 'text-plain'
  except urllib.error.HTTPError as err:
    context.http_response['status'] = str(err.code)
    context.http_response['body'] = err.read().decode('utf-8')
    context.http_response['content-type'] = 'text-plain'


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
