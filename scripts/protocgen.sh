#!/usr/bin/env bash

set -eo pipefail

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  protoc \
  -I "proto" \
  -I "third_party/proto" \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  # command to generate gRPC gateway (*.pb.gw.go in respective modules) files
  protoc \
  -I "proto" \
  -I "third_party/proto" \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  # generate swagger only for query files
  query_file=$(find "${dir}" -maxdepth 1 -name 'query.proto')
  if [[ ! -z "$query_file" ]]; then
    protoc  \
    -I "proto" \
    -I "third_party/proto" \
    "$query_file" \
    --swagger_out=logtostderr=true,stderrthreshold=1000,fqn_for_swagger_name=true,simple_operation_ids=true:.
  fi
done

# combine swagger files
# uses nodejs package `swagger-combine`.
swagger-combine ./client/grpc-gateway/config.json -o ./client/grpc-gateway/swagger.json --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
find ./ -name 'query.swagger.json' -exec rm {} \;

# generate codec/testdata proto code
protoc -I "proto" -I "third_party/proto" -I "testutil/testdata" --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. ./testutil/testdata/proto.proto

# move proto files to the right places
cp -r github.com/cosmos/cosmos-sdk/* ./
rm -rf github.com
