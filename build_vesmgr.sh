#!/bin/bash
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

# Install RMR from deb packages at packagecloud.io
rmr=rmr_4.1.2_amd64.deb
wget --content-disposition  https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/$rmr/download.deb
sudo dpkg -i $rmr
rm $rmr
rmrdev=rmr-dev_4.1.2_amd64.deb
wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/$rmrdev/download.deb
sudo dpkg -i $rmrdev
rm $rmrdev

# Required to find nng and rmr libs
export LD_LIBRARY_PATH=/usr/local/lib

# Go install, build, etc
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH

# xApp-framework stuff
export CFG_FILE=$PWD/config/config-file-ut.json
export RMR_SEED_RT=$PWD/config/uta_rtg.rt

GO111MODULE=on GO_ENABLED=0 GOOS=linux

# Run vesmgr UT
GO111MODULE=on go test -v -p 1 -cover -coverprofile=coverage.out ./...

# setup version tag
if [ -f container-tag.yaml ]
then
    tag=$(grep "tag:" container-tag.yaml | awk '{print $2}')
else
    tag="-"
fi

hash=$(git rev-parse --short HEAD || true)

# Install vespamgr
go install -ldflags "-X main.Version=$tag -X main.Hash=$hash" -v $PWD/cmd/vespamgr
