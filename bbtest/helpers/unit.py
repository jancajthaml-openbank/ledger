#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import docker
import platform
import tarfile
import tempfile
import errno
import os
import subprocess
from helpers.shell import execute


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
      "METRICS_CONTINUOUS": True,
      "METRICS_REFRESHRATE": "1s",
      "METRICS_OUTPUT": os.path.realpath('{}/../../reports/blackbox-tests/metrics'.format(os.path.dirname(__file__))),
      "STORAGE": "/data",
    }

  def get_arch(self):
    return {
      'x86_64': 'amd64',
      'armv7l': 'armhf',
      'armv8': 'arm64'
    }.get(platform.uname().machine, 'amd64')

  def __init__(self, context):
    self.arch = self.get_arch()

    self.store = dict()
    self.image_version = None
    self.debian_version = None
    self.units = list()
    self.docker = docker.from_env()
    self.context = context

  def install(self, file):
    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    (code, result, error) = execute(['dpkg', '-c', file])
    if code != 0:
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))
    else:
      os.makedirs('{}/reports/blackbox-tests/meta'.format(cwd), exist_ok=True)
      with open('{}/reports/blackbox-tests/meta/debian.ledger.txt'.format(cwd), 'w') as fd:
        fd.write(result)

      result = [item for item in result.split(os.linesep)]
      result = [item.rsplit('/', 1)[-1].strip() for item in result if "/lib/systemd/system/ledger" in item]
      result = [item for item in result if not item.endswith('unit.slice')]

      self.units = result

  def download(self):
    self.image_version = os.environ.get('IMAGE_VERSION', '')
    self.debian_version = os.environ.get('UNIT_VERSION', '')

    if self.debian_version.startswith('v'):
      self.debian_version = self.debian_version[1:]

    assert self.image_version, 'IMAGE_VERSION not provided'
    assert self.debian_version, 'UNIT_VERSION not provided'

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    self.binary = '{}/packaging/bin/ledger_{}_{}.deb'.format(cwd, self.debian_version, self.arch)

    if os.path.exists(self.binary):
      self.install(self.binary)
      return

    os.makedirs(os.path.dirname(self.binary), exist_ok=True)

    failure = None
    image = 'openbank/ledger:{}'.format(self.image_version)
    package = '/opt/artifacts/ledger_{}_{}.deb'.format(self.debian_version, self.arch)
    temp = tempfile.NamedTemporaryFile(delete=True)
    try:
      with open(temp.name, 'w') as fd:
        fd.write(str(os.linesep).join([
          'FROM alpine',
          'COPY --from={} {} {}'.format(image, package, self.binary)
        ]))

      image, stream = self.docker.images.build(fileobj=temp, rm=True, pull=False, tag='bbtest_artifacts-scratch')
      for chunk in stream:
        if not 'stream' in chunk:
          continue
        for line in chunk['stream'].splitlines():
          l = line.strip(os.linesep)
          if not len(l):
            continue
          print(l)

      scratch = self.docker.containers.run('bbtest_artifacts-scratch', ['/bin/true'], detach=True)

      tar_name = tempfile.NamedTemporaryFile(delete=True)
      with open(tar_name.name, 'wb') as fd:
        bits, stat = scratch.get_archive(self.binary)
        for chunk in bits:
          fd.write(chunk)

      archive = tarfile.TarFile(tar_name.name)
      archive.extract(os.path.basename(self.binary), os.path.dirname(self.binary))
      self.install(self.binary)
      scratch.remove()
    except Exception as ex:
      failure = ex
    finally:
      temp.close()
      try:
        self.docker.images.remove('bbtest_artifacts-scratch', force=True)
      except:
        pass

    if failure:
      raise failure

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

    (code, result, error) = execute(['journalctl', '-o', 'cat', '--no-pager'])
    if code == 0:
      with open('{}/reports/blackbox-tests/logs/journal.log'.format(cwd), 'w') as fd:
        fd.write(result)

    for unit in set(self.__get_systemd_units() + self.units):
      (code, result, error) = execute(['journalctl', '-o', 'cat', '-u', unit, '--no-pager'])
      if code != 0 or not result:
        continue
      with open('{}/reports/blackbox-tests/logs/{}.log'.format(cwd, unit), 'w') as fd:
        fd.write(result)

  def teardown(self):
    self.collect_logs()
    for unit in self.__get_systemd_units():
      execute(['systemctl', 'stop', unit])
    self.collect_logs()

  def __get_systemd_units(self):
    (code, result, error) = execute(['systemctl', 'list-units', '--no-legend'])
    result = [item.split(' ')[0].strip() for item in result.split(os.linesep)]
    result = [item for item in result if "ledger" in item and not item.endswith('unit.slice')]
    return result
