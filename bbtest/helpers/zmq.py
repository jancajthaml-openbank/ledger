#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import zmq
import threading
import time
import re


class ZMQHelper(threading.Thread):

  def __init__(self, context):
    threading.Thread.__init__(self)
    self.__cancel = threading.Event()
    self.__mutex = threading.Lock()
    self.backlog = []
    self.context = context
    self.vault_unit_message = re.compile(r'^VaultUnit\/([^\s]{1,100}) LedgerUnit\/([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) (-?\d{1,100}\.\d{1,100}|-?\d{1,100}) ([A-Z]{3})$')
    self.working = True

  def clear(self):
    self.working = True

  def silence(self):
    self.working = False

  def start(self):
    ctx = zmq.Context.instance()

    self.__pull_url = 'tcp://127.0.0.1:5562'
    self.__pub_url = 'tcp://127.0.0.1:5561'

    self.__pub = ctx.socket(zmq.PUB)
    self.__pub.bind(self.__pub_url)

    self.__pull = ctx.socket(zmq.PULL)
    self.__pull.bind(self.__pull_url)
    self.__pull.set_hwm(100)

    threading.Thread.start(self)

  def run(self):
    while not self.__cancel.is_set():
      try:
        data = self.__pull.recv(zmq.NOBLOCK)
        if not (data and self.working):
          continue
        self.__pub.send(data)
        self.__process_next_message(data)
        self.backlog.append(data)
      except zmq.error.Again as ex:
        if ex.errno != 11:
          return

  def __process_next_message(self, data):
    message = data.decode('utf-8')
    if not message.startswith('VaultUnit/'):
      return

    m = self.vault_unit_message.match(message)
    if not m:
      return

    tenant, sender, account, req_id, kind, transaction, amount, currency = m.groups()
    reply_event = self.context.vault.process_account_event(tenant, account, kind, transaction, amount, currency)
    self.__pub.send('LedgerUnit/{} VaultUnit/{} {} {} {}'.format(sender, tenant, req_id, account, reply_event).encode())

  def send(self, data):
    self.__pub.send(data.encode())

  def ack(self, data):
    self.__mutex.acquire()
    self.backlog = [item for item in self.backlog if item != data]
    self.__mutex.release()

  def stop(self):
    if self.__cancel.is_set():
      return
    self.__cancel.set()
    try:
      self.join()
    except:
      pass
    self.__pub.close()
    self.__pull.close()
