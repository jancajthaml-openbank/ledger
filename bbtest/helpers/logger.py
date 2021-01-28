#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import logging
from systemd import journal


def logger():
  log = logging.getLogger('bbtest')
  log.addHandler(journal.JournaldLogHandler())
  log.setLevel(logging.DEBUG)
  return log