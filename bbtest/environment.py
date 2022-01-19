#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
from helpers.unit import UnitHelper
from helpers.zmq import ZMQHelper
from helpers.vault import VaultHelper
from openbank_testkit import StatsdMock
from helpers.logger import logger


def before_feature(context, feature):
  context.statsd.clear()
  context.log.info('')
  context.log.info('  (FEATURE) {}'.format(feature.name))


def before_scenario(context, scenario):
  context.zmq.clear()
  context.log.info('')
  context.log.info('  (SCENARIO) {}'.format(scenario.name))
  context.log.info('')


def after_scenario(context, scenario):
  context.unit.collect_logs()


def before_all(context):
  context.log = logger()
  context.log.info('')
  context.log.info('  (START)')
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper(context)
  context.vault = VaultHelper(context)
  context.statsd = StatsdMock()
  context.statsd.start()
  context.zmq.start()
  context.unit.configure()
  context.unit.download()


def after_all(context):
  context.log.info('')
  context.log.info('  (END)')
  context.log.info('')
  context.unit.teardown()
  context.zmq.stop()
  context.statsd.stop()
