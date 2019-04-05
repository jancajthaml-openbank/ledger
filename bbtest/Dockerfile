# ---------------------------------------------------------------------------- #

# Copyright (c) 2017-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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

# ---------------------------------------------------------------------------- #

FROM jancajthaml/bbtest

COPY packaging/bin /etc/bbtest/packages

RUN \
    if [ -z "$(find . /etc/bbtest/packages -type f -name 'ledger_*_amd64.deb' -print)" ] ; then \
      (>&2 echo "no package to test found") ; \
      exit 1 ; \
    fi

RUN find . /etc/bbtest/packages -type f -name 'ledger_*_amd64.deb' | \
    xargs -I {} mv {} /etc/bbtest/packages/ledger.deb

RUN apt-get update && \
    apt-get install -y \
      init-system-helpers>=1.18~ \
      libzmq5>=4.2.1~

# ---------------------------------------------------------------------------- #
