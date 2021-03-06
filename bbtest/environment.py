#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
from helpers.unit import UnitHelper
from helpers.zmq import ZMQHelper
from helpers.vault import VaultHelper
from helpers.statsd import StatsdHelper
from helpers.logger import logger


def before_feature(context, feature):
  context.statsd.clear()
  context.log.info('')
  context.log.info('  (FEATURE) {}'.format(feature.name))


def before_scenario(context, scenario):
  context.log.info('')
  context.log.info('  (SCENARIO) {}'.format(scenario.name))
  context.log.info('')


def after_scenario(context, scenario):
  context.unit.collect_logs()


def after_feature(context, feature):
  context.zmq.clear()


def before_all(context):
  context.log = logger()
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper(context)
  context.vault = VaultHelper(context)
  context.statsd = StatsdHelper()
  context.statsd.start()
  context.zmq.start()
  context.unit.configure()
  context.unit.download()


def after_all(context):
  context.unit.teardown()
  context.zmq.stop()
  context.statsd.stop()
