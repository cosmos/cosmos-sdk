#!/usr/bin/env bash

set -eo pipefail

go install github.com/gogo/protobuf/protoc-gen-gogotypes

protoc -I. --gogotypes_out=./types third_party/proto/google/protobuf/any.proto
mv types/third_party/proto/google/protobuf/any.pb.go types
rm -rf types/third_party
