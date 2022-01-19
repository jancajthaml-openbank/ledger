#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import os


@then('transaction {transaction} of tenant {tenant} should be')
def check_transaction_state(context, tenant, transaction):
  filename = os.path.realpath('/data/t_{}/transaction/{}'.format(tenant, transaction))

  assert os.path.isfile(filename) is True, 'file {} does not exists'.format(filename)

  actual = dict()
  with open(filename, 'r') as fd:
    lines = fd.readlines()
    lines = [line.strip() for line in lines]
    state = lines[0]
    transfers = []

    for transfer in lines[1:]:
      item = transfer.split(' ')

      transfers.append({
        'id': item[0],
        'credit': {
          'name': item[2],
          'tenant': item[1],
        },
        'debit':{
          'name': item[4],
          'tenant': item[3],
        },
        'amount': item[6],
        'currency': item[7],
        'valueDate': item[5],
      })

    actual.update({
      'state': state,
      'transfers': transfers,
    })

  for row in context.table:
    assert row['key'] in actual, 'key {} not found in {}'.format(row['key'], actual)
    assert actual[row['key']] == row['value'], 'extected {} got {} instead'.format(row['key'], actual)


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
