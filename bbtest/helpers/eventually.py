#!/usr/bin/env python
# -*- coding: utf-8 -*-

import time


class eventually():

  def __init__(self, timeout=2):
    self.__inblock = False
    self.__last_exception = None
    self.__timeout = timeout
    self.__block = lambda *args: None

  def __get__(self, instance, *args):
    return partial(self.__call__, instance)

  def __call__(self, *args, **kwargs):
    if not self.__inblock:
      self.__block = args[0]
      self.__inblock = True
      return self

    deadline = time.monotonic() + self.__timeout
    while deadline > time.monotonic():
      try:
        return self.__block(*args, **kwargs)
      except AssertionError as ex:
        self.__last_exception = ex
        time.sleep(0.5)
      except Exception as ex:
        raise ex
    if self.__last_exception:
      raise self.__last_exception
