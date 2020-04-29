#!/usr/bin/env bash

set -eo pipefail

go install github.com/gogo/protobuf/protoc-gen-gogotypes

protoc -I. --gogotypes_out=./codec/types third_party/proto/google/protobuf/any.proto
mv codec/types/third_party/proto/google/protobuf/any.pb.go codec/types
rm -rf codec/types/third_party
