# Copyright (c) 2017-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
# stage 0 bootstrap dependencies

FROM amd64/debian:buster

COPY packaging/bin/* /opt/artifacts/

USER root

RUN apt-get update

RUN find /opt/artifacts -name ledger_*_amd64.deb | xargs -I{} -tx apt-get install -y -f {}

# ---------------------------------------------------------------------------- #
# stage 1 link files

FROM amd64/debian:buster

COPY --from=0 /opt/artifacts /opt/artifacts

COPY --from=0 /usr/lib/x86_64-linux-gnu /usr/lib/x86_64-linux-gnu
COPY --from=0 /lib/x86_64-linux-gnu /lib/x86_64-linux-gnu

RUN ln -s /opt/artifacts/ledger-rest-linux-amd64 /usr/bin/ledger-rest && \
    ln -s /opt/artifacts/ledger-unit-linux-amd64 /usr/bin/ledger-unit

ENTRYPOINT [ "ledger-rest" ]
