#!/bin/bash

# Copyright 2024 Google LLC
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

# Run this script from the root of the CQL repository to regenerate these
# protos.

if ! [ -x "$(command -v protoc)" ]; then
  echo 'protoc missing, please install the protobuf compiler. On linux: sudo apt install -y protobuf-compiler'
fi

if [ "${GOPATH}" = "" ]; then
  echo "Setting temp GOPATH environment variable, creating GOPATH directory \$HOME/go"
  GOPATH="${HOME}/go"
  mkdir -p $GOPATH
fi

if [ "${GOBIN}" = "" ]; then
  echo "Setting temp GOBIN environment variable, creating GOBIN directory in \$GOPATH/bin"
  GOBIN="${GOPATH}/bin"
  mkdir -p $GOBIN
fi

PROTO_TMP_DIR=/tmp/protostaging
mkdir -p $PROTO_TMP_DIR

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.29.0
PATH=$PATH:$GOBIN

GOOGLEAPIS_DIR=/tmp/googleapis
if [ ! -d "$GOOGLEAPIS_DIR" ] ; then
  mkdir -p /tmp/googleapis
  git clone https://github.com/googleapis/googleapis.git /tmp/googleapis
else
  echo "warn: skipping cloning github.com/googleapis/googleapis.git because /tmp/googleapis already exists"
fi

find protos -type f -name '*.proto' -exec protoc --proto_path=/tmp/googleapis  --proto_path=. --go_out=$PROTO_TMP_DIR {} \;

cp -R $PROTO_TMP_DIR/github.com/google/cql/protos/* protos/
