#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import ssl
import urllib.request
import urllib.response
import socket
import time
import http
from io import StringIO


class StubHeaders(object):
  def __init(self):
    pass

  def get_content_type(self):
    return 'text-plain'


class StubResponse(object):

  def __init__(self, status, body):

    self.status = status
    self.body = body
    self.headers = StubHeaders()

  def read(self):
    return self.body.read()

  def info(self):
    return self.headers


class Request(object):

  def __init__(self, **kwargs):
    self.__underlying = urllib.request.Request(**kwargs)

  def add_header(self, key, value):
    self.__underlying.add_header(key, value)

  @property
  def data(self):
    return self.__underlying.data

  @data.setter
  def data(self, value):
    self.__underlying.data = value

  def do(self):
    timeout = 10
    last_exception = None

    deadline = time.monotonic() + timeout
    while deadline > time.monotonic():
      try:
        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE
        return urllib.request.urlopen(self.__underlying, timeout=timeout, context=ctx)
      except (http.client.RemoteDisconnected, socket.timeout):
        return StubResponse(504, StringIO(''))
      except urllib.error.HTTPError as err:
        return StubResponse(err.code, err)
      except urllib.error.URLError as err:
        last_exception = err
        time.sleep(0.5)
      except ssl.SSLError as err:
        last_exception = err
        time.sleep(0.5)

    if last_exception:
      raise last_exception
    else:
      raise AssertionError('timeout')
