#
#  Copyright (c) 2019 AT&T Intellectual Property.
#  Copyright (c) 2018-2019 Nokia.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#
#   This source code is part of the near-RT RIC (RAN Intelligent Controller)
#   platform project (RICP).
#

FROM nexus3.o-ran-sc.org:10004/o-ran-sc/bldr-ubuntu18-c-go:9-u18.04 as ubuntu-vespamgr

# Install utilities
RUN apt update && apt install -y iputils-ping net-tools curl sudo

# Set the Working Directory for ves-agent inside the container
RUN mkdir -p $GOPATH/src/VESPA
WORKDIR $GOPATH/src/VESPA

# Clone VES Agent v0.3.0 from github
RUN git clone -b v0.3.0 https://github.com/nokia/ONAP-VESPA.git $GOPATH/src/VESPA

# Install VES Agent
RUN export GOPATH=$HOME/go && \
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH && \
    go install -v ./ves-agent

# Set the Working Directory for vespamgr inside the container
RUN mkdir -p $GOPATH/src/vespamgr
WORKDIR $GOPATH/src/vespamgr
COPY $HOME/ .

RUN ./build_vesmgr.sh

# Final, executable and deployable container
FROM ubuntu:18.04

RUN mkdir -p /etc/ves-agent

COPY --from=ubuntu-vespamgr /usr/local/lib /usr/local/lib
COPY --from=ubuntu-vespamgr root/go/bin /root/go/bin
COPY --from=ubuntu-vespamgr $GOPATH/src/vespamgr/config/config-file.json /etc/ves-agent/config-file.json
COPY --from=ubuntu-vespamgr $GOPATH/src/vespamgr/config/uta_rtg.rt /etc/ves-agent/uta_rtg.rt

RUN ldconfig

ENV CFG_FILE=/etc/ves-agent/config-file.json
ENV RMR_SEED_RT=/etc/ves-agent/uta_rtg.rt
ENV RMR_RTG_SVC="service-ricplt-rtmgr-rmr.ricplt:4560"

ENV PATH="/root/go/bin:${PATH}"

ENTRYPOINT ["vesmgr"]
