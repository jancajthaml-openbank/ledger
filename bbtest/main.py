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
    '-o {}/../reports/blackbox-tests/behave/results.json'.format(cwd),
  ]

  if str(os.environ.get('CI', 'false')) == 'false':
    args.append('-f plain')
    args.append('--tags=~@wip')
  else:
    args.append('-f progress3')
    args.append('--tags=~@wip')
    args.append('--quiet')

  args.append('@{}/order.txt'.format(cwd))

  for path in [
    '/data',
    '{}/../reports/blackbox-tests/metrics'.format(cwd),
    '{}/../reports/blackbox-tests/logs'.format(cwd),
    '{}/../reports/blackbox-tests/meta'.format(cwd),
    '{}/../reports/blackbox-tests/behave'.format(cwd),
    '{}/../reports/blackbox-tests/cucumber'.format(cwd),
    '{}/../reports/blackbox-tests/junit'.format(cwd)
  ]:
    os.system('mkdir -p {}'.format(path))
    os.system('rm -rf {}/*'.format(path))

  from behave import __main__ as behave_executable

  exit_code = behave_executable.main(args=' '.join(args))

  with open('{}/../reports/blackbox-tests/behave/results.json'.format(cwd), 'r') as fd_behave:
    cucumber_data = None
    with open('{}/../reports/blackbox-tests/cucumber/results.json'.format(cwd), 'w') as fd_cucumber:
      behave_data = json.loads(fd_behave.read())
      cucumber_data = json.dumps(behave2cucumber.convert(behave_data))
      fd_cucumber.write(cucumber_data)

  execute([
    'json_to_junit',
    '{}/../reports/blackbox-tests/cucumber/results.json'.format(cwd),
    '{}/../reports/blackbox-tests/junit/results.xml'.format(cwd)
  ])

  sys.exit(exit_code)
