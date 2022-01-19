#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import os
from behave import *
from openbank_testkit import Shell, Request
from helpers.eventually import eventually


@given('package {package} is {operation}')
def step_impl(context, package, operation):
  if operation == 'installed':
    (code, result, error) = Shell.run(["apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confold", context.unit.binary])
    assert code == 'OK', "unable to install with code {} and {} {}".format(code, result, error)
    assert os.path.isfile('/etc/ledger/conf.d/init.conf') is True, 'config file does not exists'
    Shell.run(['systemctl', 'start', package])
  elif operation == 'uninstalled':
    (code, result, error) = Shell.run(["apt-get", "-f", "-qq", "remove", package])
    assert code == 'OK', "unable to uninstall with code {} and {} {}".format(code, result, error)
    (code, result, error) = Shell.run(["apt-get", "-f", "-qq", "purge", package])
    assert code == 'OK', "unable to purge with code {} and {} {}".format(code, result, error)
    assert os.path.isfile('/etc/ledger/conf.d/init.conf') is False, 'config file still exists'
  else:
    assert False, 'unknown operation {}'.format(operation)


@given('systemctl contains following active units')
@then('systemctl contains following active units')
def step_impl(context):
  (code, result, error) = Shell.run(["systemctl", "list-units", "--all", "--no-legend", "--state=active"])
  assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)
  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])
  result = [item.replace('*', '').strip().split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if item in items]
  assert len(result) > 0, 'units not found\n{}'.format(result)


@given('systemctl does not contain following active units')
@then('systemctl does not contain following active units')
def step_impl(context):
  (code, result, error) = Shell.run(["systemctl", "list-units", "--all", "--no-legend", "--state=active"])
  assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)
  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])
  result = [item.replace('*', '').strip().split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if item in items]
  assert len(result) == 0, 'units found\n{}'.format(result)


@given('unit "{unit}" is running')
@then('unit "{unit}" is running')
def unit_running(context, unit):
  @eventually(10)
  def wait_for_unit_state_change():
    (code, result, error) = Shell.run(["systemctl", "show", "-p", "SubState", unit])
    assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)
    assert 'SubState=running' in result, result

  wait_for_unit_state_change()

  if 'ledger-rest' in unit:
    request = Request(method='GET', url="https://127.0.0.1/health")

    @eventually(5)
    def wait_for_healthy():
      response = request.do()
      assert response.status == 200, str(response.status)

    wait_for_healthy()


@given('unit "{unit}" is not running')
@then('unit "{unit}" is not running')
def unit_not_running(context, unit):
  @eventually(10)
  def wait_for_unit_state_change():
    (code, result, error) = Shell.run(["systemctl", "show", "-p", "SubState", unit])
    assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)
    assert 'SubState=dead' in result, result
  wait_for_unit_state_change()


@given('{operation} unit "{unit}"')
@when('{operation} unit "{unit}"')
def operation_unit(context, operation, unit):
  (code, result, error) = Shell.run(["systemctl", operation, unit])
  assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)


@given('{unit} is configured with')
def unit_is_configured(context, unit):
  params = dict()
  for row in context.table:
    params[row['property']] = row['value'].strip()
  context.unit.configure(params)


@then('tenant {tenant} is offboarded')
@given('tenant {tenant} is offboarded')
def offboard_unit(context, tenant):
  logfile = os.path.realpath('{}/../../reports/blackbox-tests/logs/ledger-unit.{}.log'.format(os.path.dirname(__file__), tenant))
  (code, result, error) = Shell.run(['journalctl', '-o', 'cat', '-u', 'ledger-unit@{}.service'.format(tenant), '--no-pager'])
  if code == 'OK' and result:
    with open(logfile, 'w') as f:
      f.write(result)
  Shell.run(['systemctl', 'stop', 'ledger-unit@{}.service'.format(tenant)])
  (code, result, error) = Shell.run(['journalctl', '-o', 'cat', '-u', 'ledger-unit@{}.service'.format(tenant), '--no-pager'])
  if code == 'OK' and result:
    with open(logfile, 'w') as fd:
      fd.write(result)
  Shell.run(['systemctl', 'disable', 'ledger-unit@{}.service'.format(tenant)])
  unit_not_running(context, 'ledger-unit@{}'.format(tenant))


@given('tenant {tenant} is onboarded')
def onboard_unit(context, tenant):
  (code, result, error) = Shell.run(["systemctl", 'enable', 'ledger-unit@{}'.format(tenant)])
  assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)
  (code, result, error) = Shell.run(["systemctl", 'start', 'ledger-unit@{}'.format(tenant)])
  assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)
  unit_running(context, 'ledger-unit@{}'.format(tenant))
