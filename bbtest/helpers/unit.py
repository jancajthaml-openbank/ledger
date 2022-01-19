#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
from openbank_testkit import Shell, Package, Platform


class UnitHelper(object):

  @staticmethod
  def default_config():
    return {
      "LOG_LEVEL": "DEBUG",
      "HTTP_PORT": 443,
      "SERVER_KEY": "/etc/ledger/secrets/domain.local.key",
      "SERVER_CERT": "/etc/ledger/secrets/domain.local.crt",
      "LAKE_HOSTNAME": "127.0.0.1",
      "TRANSACTION_INTEGRITY_SCANINTERVAL": "24h",
      "MEMORY_THRESHOLD": 0,
      "STORAGE_THRESHOLD": 0,
      "STATSD_ENDPOINT": "127.0.0.1:8125",
      "STORAGE": "/data",
    }

  def __init__(self, context):
    self.store = dict()
    self.units = list()
    self.context = context

  def download(self):
    version = os.environ.get('VERSION', '')
    meta = os.environ.get('META', '')

    if version.startswith('v'):
      version = version[1:]

    assert version, 'VERSION not provided'
    assert meta, 'META not provided'

    package = Package('ledger')

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    assert package.download(version, meta, '{}/packaging/bin'.format(cwd)), 'unable to download package ledger'

    self.binary = '{}/packaging/bin/ledger_{}_{}.deb'.format(cwd, version, Platform.arch)


  def configure(self, params = None):
    options = dict()
    options.update(UnitHelper.default_config())
    if params:
      options.update(params)

    os.makedirs('/etc/ledger/conf.d', exist_ok=True)
    with open('/etc/ledger/conf.d/init.conf', 'w') as fd:
      fd.write(str(os.linesep).join("LEDGER_{!s}={!s}".format(k, v) for (k, v) in options.items()))

  def collect_logs(self):
    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    os.makedirs('{}/reports/blackbox-tests/logs'.format(cwd), exist_ok=True)

    (code, result, error) = Shell.run(['journalctl', '-o', 'cat', '--no-pager'])
    if code == 'OK':
      with open('{}/reports/blackbox-tests/logs/journal.log'.format(cwd), 'w') as fd:
        fd.write(result)

    for unit in set(self.__get_systemd_units() + self.units):
      (code, result, error) = Shell.run(['journalctl', '-o', 'cat', '-u', unit, '--no-pager'])
      if code != 'OK' or not result:
        continue
      with open('{}/reports/blackbox-tests/logs/{}.log'.format(cwd, unit), 'w') as fd:
        fd.write(result)

  def teardown(self):
    self.collect_logs()
    for unit in self.__get_systemd_units():
      Shell.run(['systemctl', 'stop', unit])
    self.collect_logs()

  def __get_systemd_units(self):
    (code, result, error) = Shell.run(['systemctl', 'list-units', '--all', '--no-legend'])
    result = [item.replace('*', '').strip().split(' ')[0].strip() for item in result.split(os.linesep)]
    result = [item for item in result if "ledger" in item and not item.endswith('unit.slice')]
    return result
