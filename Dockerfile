# Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM alpine

COPY packaging/bin/* /opt/artifacts/

RUN \
    if \
      \
      [ ! -f /opt/artifacts/ledger-rest-linux-amd64 ] || \
      [ ! -f /opt/artifacts/ledger-unit-linux-amd64 ] || \
      [ ! -f /opt/artifacts/ledger-rest-linux-armhf ] || \
      [ ! -f /opt/artifacts/ledger-unit-linux-armhf ] || \
      \
      [ -z "$(find /opt/artifacts -type f -name 'ledger_*_amd64.deb' -print)" ] || \
      [ -z "$(find /opt/artifacts -type f -name 'ledger_*_armhf.deb' -print)" ] \
      \
      ; then \
      (>&2 echo "missing expected files, run package and debian for both amd64 and armhf") ; \
      exit 1 ; \
    fi

ENTRYPOINT [ "echo", "only stores candidate binaries in /opt/artifacts" ]
