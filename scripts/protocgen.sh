#!/usr/bin/env bash

set -eo pipefail

proto_dirs=$(find . -path ./vendor -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  protoc -I/usr/local/include \
  -I. \
  -I${GOPATH}/src \
  -Ivendor/github.com/gogo/protobuf/gogoproto \
  --gofast_out=$GOPATH/src \
  $(find "${dir}" -name '*.proto')
done
