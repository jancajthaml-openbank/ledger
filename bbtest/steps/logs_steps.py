#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import re
import os
from behave import *
from helpers.shell import execute
from helpers.eventually import eventually


@then('journalctl of "{unit}" contains following')
def step_impl(context, unit):
  expected_lines = [item.strip() for item in context.text.split('\n') if len(item.strip())]
  ansi_escape = re.compile(r'(?:\x1B[@-_]|[\x80-\x9F])[0-?]*[ -/]*[@-~]')

  def get_unit_description():
    (code, result, error) = execute(["systemctl", "status", unit])
    result = ansi_escape.sub('', result)
    assert len(result), str(result) + ' ' + str(error)
    result = result.split(os.linesep)[0]
    pivot = "{} - ".format(unit)
    idx = result.rfind(pivot)
    if idx >= 0:
      return result[idx+len(pivot):]
    else:
      return None

  description = get_unit_description()

  @eventually(5)
  def impl():
    (code, result, error) = execute(['journalctl', '-o', 'cat', '-u', unit, '--no-pager'])
    result = ansi_escape.sub('', result)
    assert code == 'OK', str(result) + ' ' + str(error)

    if description:
      idx = result.rfind("Stopped {}.".format(description))
    else:
      idx = -1

    if idx > 0:
      result = result[idx:]

    actual_lines = [item.strip() for item in result.split('\n') if len(item.strip())]

    for expected in expected_lines:
      found = False
      for actual in actual_lines:
        if expected in actual:
          found = True
          break

      assert found, 'message "{}" was not found in logs\n  {}'.format(context.text.strip(), '\n  '.join(actual_lines))

  impl()
