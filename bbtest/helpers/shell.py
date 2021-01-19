#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import subprocess
import threading
import signal
import time
import os
import re


class Deadline(threading.Thread):

  def __init__(self, timeout, callback):
    super().__init__(daemon=True)
    self.__timeout = timeout
    self.__callback = callback
    self.__cancelled = threading.Event()

  def run(self) -> None:
    deadline = time.monotonic() + self.__timeout
    while not self.__cancelled.wait(deadline - time.monotonic()):
      if not self.__cancelled.is_set() and deadline <= time.monotonic():
        return self.__callback()

  def cancel(self) -> None:
    self.__cancelled.set()
    self.join()


def execute(command, timeout=60) -> None:
  ansi_escape = re.compile(r'(\x9B|\x1B\[)[0-?]*[ -/]*[@-~]', flags=re.IGNORECASE)
  try:
    p = subprocess.Popen(
      command,
      shell=False,
      stdin=None,
      stdout=subprocess.PIPE,
      stderr=subprocess.PIPE,
      close_fds=True
    )

    def kill() -> None:
      for sig in [signal.SIGTERM, signal.SIGQUIT, signal.SIGKILL, signal.SIGKILL]:
        if p.poll():
          break
        try:
          os.kill(p.pid, sig)
        except OSError:
          break

    deadline = Deadline(timeout, callback=kill)
    deadline.start()
    (result, error) = p.communicate()
    deadline.cancel()

    result = result.decode('utf-8').strip() if result else ''
    result = ansi_escape.sub('', result)
    error = error.decode('utf-8').strip() if error else ''
    error = ansi_escape.sub('', error)
    code = p.returncode

    del p

    return (code, result, error)
  except subprocess.CalledProcessError:
    return (-1, None, None)
