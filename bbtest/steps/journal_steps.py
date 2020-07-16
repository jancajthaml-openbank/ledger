from behave import *
import ssl
import urllib.request
import json
import os
import glob
from datetime import datetime


@then('transaction {transaction} of tenant {tenant} should be')
def check_transaction_state(context, tenant, transaction):

  path = '/data/t_{}/transaction/{}'.format(tenant, transaction)
  assert os.path.isfile(path) is True, 'file {} does not exists'.format(path)

  actual = dict()
  with open(path, 'r') as fd:
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
    assert row['key'] in actual
    assert actual[row['key']] == row['value']
