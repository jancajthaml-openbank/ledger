#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import ssl
import urllib.request
import urllib.response
import socket
import time
import http
from io import BytesIO


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
    self.__underlying.data = value.encode('utf-8')

  def do(self):
    timeout = 30
    last_exception = None
    
    ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    ctx.check_hostname = False
    ctx.options |= (
      ssl.OP_NO_SSLv2 | ssl.OP_NO_SSLv3 | ssl.OP_NO_TLSv1 | ssl.OP_NO_TLSv1_1
    )
    ctx.set_ciphers("ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM:DHE+CHACHA20")
    ctx.verify_mode = ssl.CERT_NONE

    deadline = time.monotonic() + (2 * timeout)
    while deadline > time.monotonic():
      try:
        return urllib.request.urlopen(self.__underlying, timeout=timeout, context=ctx)
      except (http.client.RemoteDisconnected, socket.timeout):
        return StubResponse(504, BytesIO())
      except urllib.error.HTTPError as err:
        return StubResponse(err.code, err)
      except ssl.SSLError:
        return StubResponse(504, BytesIO())
      except urllib.error.URLError as err:
        last_exception = err
        time.sleep(0.5)

    if last_exception:
      raise last_exception
    else:
      raise AssertionError('timeout after {} seconds'.format(timeout))
