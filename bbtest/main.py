#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import json
import behave2cucumber
from helpers.shell import execute


if __name__ == "__main__":

  cwd = os.path.dirname(os.path.abspath(__file__))

  args = [
    '--color',
    '--no-capture',
    '--no-junit',
    '-f json',
    '-o reports/blackbox-tests/behave/results.json',
  ]

  if sys.stdout.isatty() and (str(os.environ.get('CI', 'false')) == 'false'):
    args.append('-f pretty')
  else:
    args.append('-f progress3')
    args.append('--quiet')

  args.append('@{}/order.txt'.format(cwd))

  for path in [
    'reports/blackbox-tests/metrics',
    'reports/blackbox-tests/logs',
    'reports/blackbox-tests/meta',
    'reports/blackbox-tests/data',
    'reports/blackbox-tests/behave',
    'reports/blackbox-tests/cucumber',
    'reports/blackbox-tests/junit'
  ]:
    os.system('mkdir -p {}'.format(path))
    os.system('rm -rf {}/*'.format(path))

  from behave import __main__ as behave_executable

  exit_code = behave_executable.main(args=' '.join(args))

  with open('reports/blackbox-tests/behave/results.json', 'r') as fd_behave:
    cucumber_data = None
    with open('reports/blackbox-tests/cucumber/results.json', 'w') as fd_cucumber:
      behave_data = json.loads(fd_behave.read())
      cucumber_data = json.dumps(behave2cucumber.convert(behave_data))
      fd_cucumber.write(cucumber_data)

  execute([
    'json_to_junit',
    'reports/blackbox-tests/cucumber/results.json',
    'reports/blackbox-tests/junit/results.xml'
  ])

  sys.exit(exit_code)
