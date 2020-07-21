
import os
from helpers.unit import UnitHelper
from helpers.zmq import ZMQHelper
from helpers.vault import VaultHelper


def after_feature(context, feature):
  context.unit.cleanup()


def before_all(context):
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper(context)
  context.vault = VaultHelper(context)
  os.system('rm -rf /tmp/reports/blackbox-tests/logs/*.log /tmp/reports/blackbox-tests/metrics/*.json /tmp/reports/blackbox-tests/data/*')
  context.zmq.start()
  context.unit.download()
  context.unit.configure()


def after_all(context):
  context.unit.teardown()
  context.zmq.stop()
  if os.path.isdir('/data'):
    os.system('cp -r /data/* /tmp/reports/blackbox-tests/data/')
