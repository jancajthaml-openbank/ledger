#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import socket
import threading
import time
import re


class StatsdHelper(threading.Thread):

  def __init__(self):
    threading.Thread.__init__(self)
    self.__cancel = threading.Event()
    self.__backlog = {}

  def start(self):
    self._sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    threading.Thread.start(self)

  def run(self):
    self._sock.bind(('127.0.0.1', 8125))

    while not self.__cancel.is_set():
      data, addr = self._sock.recvfrom(1024)
      try:
        self.process(data.decode('utf-8'))
      except:
        return

  def get(self):
    return self.__backlog

  def process(self, data):
    for metric in data.split('\n'):
      match = re.match('\A([^:]+):([^|]+)\|(.+)', metric)
      if match == None:
        continue

      key   = match.group(1)
      value = match.group(2)
      rest  = match.group(3).split('|')
      mtype = rest.pop(0)

      if (mtype == 'ms'):
        key = key + '.timer'
        if not key in self.__backlog:
          self.__backlog[key] = 0
        self.__backlog[key] = int(value)
      elif (mtype == 'g'):
        key = key + '.gauce'
        if not key in self.__backlog:
          self.__backlog[key] = 0
        self.__backlog[key] = int(value)
      elif (mtype == 'c'):
        key = key + '.count'
        if not key in self.__backlog:
          self.__backlog[key] = 0
        self.__backlog[key] += int(value)
      else:
        continue

  def stop(self):
    if self.__cancel.is_set():
      return
    self.__cancel.set()
    try:
      self._sock.shutdown(socket.SHUT_RD)
    except:
      pass
    try:
      self.join()
    except:
      pass
