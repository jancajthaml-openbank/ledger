#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import re
import os
from systemd import journal
from behave import *
from openbank_testkit import Shell
from helpers.eventually import eventually


@then('journalctl of "{unit}" contains following')
def step_impl(context, unit):
  expected_lines = [item.strip() for item in context.text.split('\n') if len(item.strip())]
  ansi_escape = re.compile(r'(?:\x1B[@-_]|[\x80-\x9F])[0-?]*[ -/]*[@-~]')
  
  actual_lines = []

  reader = journal.Reader()
  reader.this_boot()
  reader.log_level(journal.LOG_DEBUG)
  reader.add_match(_SYSTEMD_UNIT=unit)

  @eventually(5)
  def impl():
    for entry in reader:
      actual_lines.append(entry['MESSAGE'])

    for expected in expected_lines:
      found = False
      for actual in actual_lines:
        if expected in actual:
          found = True
          break

      assert found, 'message "{}" was not found in logs\n  {}'.format(context.text.strip(), '\n  '.join(actual_lines))

  impl()
