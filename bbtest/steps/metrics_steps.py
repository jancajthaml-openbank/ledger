#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import os
import stat
import json
from helpers.eventually import eventually


@then('metrics reports')
def step_impl(context):
  @eventually(10)
  def wait_for_metrics_update():
    actual = context.statsd.get()
    for row in context.table:
      key = row['key'] + '.' + row['type']
      assert key in actual, 'key {} not found in metrics'.format(key)
      if not len(row['value']):
        continue
      assert str(actual[key]) == row['value'], 'metrics {} value mismatch expected {} actual {}'.format(key, row['value'], str(actual[key]))
  wait_for_metrics_update()


