#!/usr/bin/env bash

set -eo pipefail

go install github.com/gogo/protobuf/protoc-gen-gogotypes

protoc -I. --gogotypes_out=./codec third_party/proto/google/protobuf/any.proto
mkdir -p codec/types
mv codec/third_party/proto/google/protobuf/any.pb.go codec/types
rm -rf codec/third_party
