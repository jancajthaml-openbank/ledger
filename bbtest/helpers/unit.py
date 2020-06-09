#!/usr/bin/env python
# -*- coding: utf-8 -*-

import docker
from helpers.shell import execute
import platform
import tarfile
import tempfile
import errno
import os
import subprocess


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

    self.store = {}
    self.image_version = None
    self.debian_version = None
    self.units = {}
    self.services = []
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

    scratch_docker_cmd = ['FROM alpine']

    image = 'openbank/ledger:{}'.format(self.image_version)
    package = 'ledger_{}_{}'.format(self.debian_version, self.arch)
    scratch_docker_cmd.append('COPY --from={} /opt/artifacts/{}.deb /tmp/packages/ledger.deb'.format(image, package))

    temp = tempfile.NamedTemporaryFile(delete=True)
    try:
      with open(temp.name, 'w') as f:
        for item in scratch_docker_cmd:
          f.write("%s\n" % item)

      for chunk in self.docker.build(fileobj=temp, rm=True, decode=True, tag='bbtest_artifacts-scratch'):
        if 'stream' in chunk:
          for line in chunk['stream'].splitlines():
            if len(line):
              print(line.strip('\r\n'))

      scratch = self.docker.create_container('bbtest_artifacts-scratch', '/bin/true')

      if scratch['Warnings']:
        raise Exception(scratch['Warnings'])

      tar_name = tempfile.NamedTemporaryFile(delete=True)

      tar_stream, stat = self.docker.get_archive(scratch['Id'], '/tmp/packages/ledger.deb')
      with open(tar_name.name, 'wb') as destination:
        for chunk in tar_stream:
          destination.write(chunk)

      archive = tarfile.TarFile(tar_name.name)
      archive.extract('ledger.deb', '/tmp/packages')

      (code, result, error) = execute([
        'dpkg', '-c', '/tmp/packages/ledger.deb'
      ])

      if code != 0:
        raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

      self.docker.remove_container(scratch['Id'])
    finally:
      temp.close()
      self.docker.remove_image('bbtest_artifacts-scratch', force=True)

  def configure(self, params = None):
    options = dict()
    options.update(UnitHelper.default_config())
    if params:
      options.update(params)

    with open('/etc/init/ledger.conf', 'w') as fd:
      for k, v in sorted(options.items()):
        fd.write('LEDGER_{}={}\n'.format(k, v))

  def cleanup(self):
    (code, result, error) = execute([
      'systemctl', 'list-units', '--no-legend'
    ])
    result = [item.split(' ')[0].strip() for item in result.split('\n')]
    result = [item.split('.service')[0] for item in result if ("ledger" in item and ".service" in item)]

    for unit in result:
      service = unit.split('.service')[0].split('@')[0]
      (code, result, error) = execute([
        'journalctl', '-o', 'cat', '-t', service, '-u', unit, '--no-pager'
      ])
      if code != 0 or not result:
        continue
      with open('/tmp/reports/blackbox-tests/logs/{}.log'.format(unit), 'w') as f:
        f.write(result)

  def teardown(self):
    (code, result, error) = execute([
      'systemctl', 'list-units', '--no-legend'
    ])
    result = [item.split(' ')[0].strip() for item in result.split('\n')]
    result = [item for item in result if "ledger" in item]

    for unit in result:
      execute(['systemctl', 'stop', unit])

    self.cleanup()
