#!/usr/bin/env bash

set -eo pipefail

protoc_gen_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the cosmos-sdk folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
}

protoc_gen_doc() {
  go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc 2>/dev/null
}

protoc_gen_gocosmos2() {
  cd protoc-gen-go-cosmos2
  go install .
  cd ..
}

protoc_gen_gocosmos
protoc_gen_gocosmos2
protoc_gen_doc
go mod tidy

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  buf protoc \
  -I "proto" \
  -I "third_party/proto" \
  --gocosmos_out=\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  buf protoc \
  -I "proto" \
  -I "third_party/proto" \
  --go-cosmos2_out=\
google/protobuf/duration.proto=github.com/gogo/protobuf/types,\
google/protobuf/struct.proto=github.com/gogo/protobuf/types,\
google/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
google/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
google/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  # command to generate gRPC gateway (*.pb.gw.go in respective modules) files
  buf protoc \
  -I "proto" \
  -I "third_party/proto" \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  # get the module name, e.g. from "./proto/regen/data/v1alpha1", extract "data"
  module=$(basename $(dirname $dir))

  mkdir -p ./docs/modules/${module}

  # command to generate docs using protoc-gen-doc
  buf protoc \
  -I "proto" \
  -I "third_party/proto" \
  --doc_out=./docs/modules/${module} \
  --doc_opt=docs/markdown.tmpl,protobuf.md \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

done

# generate codec/testdata proto code
buf protoc -I "proto" -I "third_party/proto" -I "testutil/testdata" --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. ./testutil/testdata/*.proto

# move proto files to the right places
cp -r github.com/cosmos/cosmos-sdk/* ./
rm -rf github.com