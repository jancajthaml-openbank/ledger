#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import os
import stat
import json
from helpers.eventually import eventually


def file_should_exist(context, path):
  @eventually(2)
  def wait_for_file_existence():
    assert os.path.isfile(path) is True, 'file {} does not exists'.format(path)
  wait_for_file_existence()


@then('metrics file {path} has permissions {permissions}')
def step_impl(context, path, permissions):
  filename = os.path.realpath('{}/../../{}'.format(os.path.dirname(__file__), path))

  file_should_exist(context, filename)
  actual = stat.filemode(os.stat(filename).st_mode)
  assert actual == permissions, "permission of {} expected {} actual {}".format(filename, permissions, actual)


@then('metrics file {path} should have following keys')
def step_impl(context, path):
  filename = os.path.realpath('{}/../../{}'.format(os.path.dirname(__file__), path))

  expected = []
  for row in context.table:
    expected.append(row['key'])
  expected = sorted(expected)
  file_should_exist(context, filename)
  with open(filename, 'r') as fd:
    assert expected == sorted(json.loads(fd.read()).keys())


@then('metrics file {path} reports')
def step_impl(context, path):
  filename = os.path.realpath('{}/../../{}'.format(os.path.dirname(__file__), path))

  file_should_exist(context, filename)

  @eventually(20)
  def wait_for_metrics_update():
    actual = dict()
    with open(filename, 'r') as fd:
      actual.update(json.loads(fd.read()))
    for row in context.table:
      assert row['key'] in actual, 'key {} not found in metrics'.format(row['key'])
      assert str(actual[row['key']]) == row['value'], 'metrics {} value mismatch expected {} actual {}'.format(row['key'], row['value'], str(actual[row['key']]))
  wait_for_metrics_update()

