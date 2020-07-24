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
      "STORAGE": "/data",
      "LOG_LEVEL": "DEBUG",
      "HTTP_PORT": 443,
      "SECRETS": "/opt/ledger/secrets",
      "LAKE_HOSTNAME": "localhost",
      "TRANSACTION_INTEGRITY_SCANINTERVAL": "24h",
      "MEMORY_THRESHOLD": 0,
      "STORAGE_THRESHOLD": 0,
      "METRICS_REFRESHRATE": "1s",
      "METRICS_OUTPUT": "/tmp/reports/blackbox-tests/metrics",
      #"METRICS_CONTINUOUS": "true",  # fixme implement
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
    self.units = dict()
    self.services = list()
    self.docker = docker.APIClient(base_url='unix://var/run/docker.sock')
    self.context = context

  def download(self):
    try:
      os.mkdir("/tmp/packages")
    except OSError as exc:
      if exc.errno != errno.EEXIST:
        raise
      pass

    self.image_version = os.environ.get('IMAGE_VERSION', '')
    self.debian_version = os.environ.get('UNIT_VERSION', '')

    if self.debian_version.startswith('v'):
      self.debian_version = self.debian_version[1:]

    image = 'openbank/ledger:{}'.format(self.image_version)
    package = '/opt/artifacts/ledger_{}_{}.deb'.format(self.debian_version, self.arch)
    target = '/tmp/packages/ledger.deb'

    temp = tempfile.NamedTemporaryFile(delete=True)
    try:
      with open(temp.name, 'w') as fd:
        fd.write(str(os.linesep).join([
          'FROM alpine',
          'COPY --from={} {} {}'.format(image, package, target)
        ]))

      for chunk in self.docker.build(fileobj=temp, rm=True, pull=False, decode=True, tag='bbtest_artifacts-scratch'):
        if not 'stream' in chunk:
          continue
        for line in chunk['stream'].splitlines():
          l = line.strip(os.linesep)
          if not len(l):
            continue
          print(l)

      scratch = self.docker.create_container('bbtest_artifacts-scratch', '/bin/true')

      if scratch['Warnings']:
        raise Exception(scratch['Warnings'])

      tar_name = tempfile.NamedTemporaryFile(delete=True)
      with open(tar_name.name, 'wb') as destination:
        tar_stream, stat = self.docker.get_archive(scratch['Id'], target)
        for chunk in tar_stream:
          destination.write(chunk)

      archive = tarfile.TarFile(tar_name.name)
      archive.extract(os.path.basename(target), os.path.dirname(target))

      (code, result, error) = execute(['dpkg', '-c', target])
      if code != 0:
        raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))
      else:
        with open('/tmp/reports/blackbox-tests/meta/debian.ledger.txt', 'w') as fd:
          fd.write(result)

      self.docker.remove_container(scratch['Id'])
    finally:
      temp.close()
      self.docker.remove_image('bbtest_artifacts-scratch', force=True)

  def configure(self, params = None):
    options = dict()
    options.update(UnitHelper.default_config())
    if params:
      options.update(params)

    os.makedirs("/etc/ledger/conf.d", exist_ok=True)
    with open('/etc/ledger/conf.d/init.conf', 'w') as fd:
      fd.write(str(os.linesep).join("LEDGER_{!s}={!s}".format(k, v) for (k, v) in options.items()))

  def cleanup(self):
    for unit in self.__get_systemd_units():
      (code, result, error) = execute(['journalctl', '-o', 'cat', '-u', unit, '--no-pager'])
      if code != 0 or not result:
        continue
      with open('/tmp/reports/blackbox-tests/logs/{}.log'.format(unit), 'w') as fd:
        fd.write(result)

  def teardown(self):
    for unit in self.__get_systemd_units():
      execute(['systemctl', 'stop', unit])
    self.cleanup()

  def __get_systemd_units(self):
    (code, result, error) = execute(['systemctl', 'list-units', '--no-legend'])
    result = [item.split(' ')[0].strip() for item in result.split(os.linesep)]
    result = [item for item in result if "ledger" in item]
    return result
