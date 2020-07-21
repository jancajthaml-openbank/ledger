#!/usr/bin/env python
# -*- coding: utf-8 -*-

from behave import *

import os
import stat
import json
from helpers.eventually import eventually


@then('file {path} should exist')
def file_should_exist(context, path):
  @eventually(2)
  def impl():
    assert os.path.isfile(path) is True, 'file {} does not exists'.format(path)
  impl()


@then('metrics file {path} has permissions {permissions}')
def step_impl(context, path, permissions):
  file_should_exist(context, path)
  assert stat.filemode(os.stat(path).st_mode) == permissions


@then('metrics file {path} should have following keys')
def step_impl(context, path):
  expected = []
  for row in context.table:
    expected.append(row['key'])
  expected = sorted(expected)

  file_should_exist(context, path)

  actual = []
  with open(path, 'r') as fd:
    actual = sorted(json.loads(fd.read()).keys())

  assert expected == actual


@then('metrics file {path} reports')
def step_impl(context, path):
  file_should_exist(context, path)

  @eventually(5)
  def wait_for_metrics_update():

    actual = dict()
    with open(path, 'r') as fd:
      actual.update(json.loads(fd.read()))

    for row in context.table:
      assert row['key'] in actual
      assert str(actual[row['key']]) == row['value'], 'key {} expected {} actual {}'.format(row['key'], row["value"], str(actual[row['key']]))

  wait_for_metrics_update()
