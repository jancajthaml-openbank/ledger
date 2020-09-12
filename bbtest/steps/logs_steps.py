#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import re
from behave import *
from helpers.shell import execute
from helpers.eventually import eventually


@then('journalctl of "{unit}" contains following')
def step_impl(context, unit):
  expected_lines = [item.strip() for item in context.text.split('\n') if len(item.strip())]
  ansi_escape = re.compile(r'(?:\x1B[@-_]|[\x80-\x9F])[0-?]*[ -/]*[@-~]')

  @eventually(5)
  def impl():
    (code, result, error) = execute(['journalctl', '-o', 'cat', '-u', unit, '--no-pager'])
    result = ansi_escape.sub('', result)
    assert code == 0, str(result) + ' ' + str(error)

    actual_lines_merged = [item.strip() for item in result.split('\n') if len(item.strip())]
    actual_lines = []
    idx = len(actual_lines_merged) - 1

    while True:
      if idx < 0 or (">>> Start <<<" in actual_lines_merged[idx]):
        break
      actual_lines.append(actual_lines_merged[idx])
      idx -= 1

    actual_lines = reversed(actual_lines)

    for expected in expected_lines:
      found = False
      for actual in actual_lines:
        if expected in actual:
          found = True
          break

      assert found == True, 'message "{}" was not found in logs'.format(context.text.strip())

  impl()
