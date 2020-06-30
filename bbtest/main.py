#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import sys
import json
import behave2cucumber

if __name__ == "__main__":

  cwd = os.path.dirname(os.path.abspath(__file__))

  args = [
    "--color",
    "--no-capture",
    "--no-junit",
    "-f json -o /tmp/reports/blackbox-tests/behave/results.json",
  ]

  if sys.stdout.isatty() and (int(os.environ.get('NO_TTY', 0)) == 0):
    args.append("-f pretty")
  else:
    args.append("-f progress3")
    args.append("--quiet")

  args.append('@{}/order.txt'.format(cwd))

  os.system('mkdir -p /tmp/reports/blackbox-tests /tmp/reports/blackbox-tests/behave /tmp/reports/blackbox-tests/cucumber')

  from behave import __main__ as behave_executable
  exit_code = behave_executable.main(args=' '.join(args))

  with open('/tmp/reports/blackbox-tests/behave/results.json', 'r') as fd_behave:
    with open('/tmp/reports/blackbox-tests/cucumber/results.json', 'w') as fd_cucumber:
      behave_data = json.loads(fd_behave.read())
      cucumber_data = json.dumps(behave2cucumber.convert(behave_data))
      fd_cucumber.write(cucumber_data)

  sys.exit(exit_code)
