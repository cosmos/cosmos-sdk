#!/usr/bin/env bash

set -eo pipefail

proto_files=''
query_files=''

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

  query_file=$(find "${dir}" -maxdepth 1 -name 'query.proto')
  if [[ ! -z "$query_file" ]]; then
    query_files=${query_files}" ${query_file}"
    protoc  \
    -I "proto" \
    -I "third_party/proto" \
    "$query_file" \
    --swagger_out=logtostderr=true,stderrthreshold=1000,fqn_for_swagger_name=true,simple_operation_ids=true:.
  fi

  proto_files=${proto_files}" ${dir:2}/*.proto"
done

#echo $query_files
#protoc  \
#-I "proto" \
#-I "third_party/proto" \
#$query_files \
#--swagger_out=logtostderr=true,include_package_in_tags=true,fqn_for_swagger_name=true,allow_merge=true:.

# generate codec/testdata proto code
protoc -I "proto" -I "third_party/proto" -I "testutil/testdata" --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. ./testutil/testdata/proto.proto

# move proto files to the right places
cp -r github.com/cosmos/cosmos-sdk/* ./
rm -rf github.com

swagger-combine ./client/grpc-gateway/config.json -o ./client/grpc-gateway/swagger.json --continueOnConflictingPaths true --includeDefinitions true --continueOnError true
