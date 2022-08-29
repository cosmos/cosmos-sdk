#!/usr/bin/env bash

# How to run manually:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen-any.sh

# This script generates a custom wrapper for google.protobuf.Any in
# codec/types/any.pb.go with a custom generated struct that lives in
# codec/types/any.go

set -eo pipefail

go install github.com/cosmos/gogoproto/protoc-gen-gogotypes
buf export buf.build/cosmos/gogo-proto --output ./third_party/proto
buf alpha protoc -I "third_party/proto" --gogotypes_out=./codec/types third_party/proto/google/protobuf/any.proto
mv codec/types/google/protobuf/any.pb.go codec/types
rm -rf codec/types/third_party

# This removes the call to RegisterType in the custom generated Any wrapper
# so that only the Any type provided by gogo protobuf is registered in the
# global gogo protobuf type registry, which we aren't actually using
sed '/proto\.RegisterType/d' codec/types/any.pb.go > tmp && mv tmp codec/types/any.pb.go
