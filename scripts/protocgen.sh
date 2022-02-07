#!/usr/bin/env bash

# Requirements
# + https://github.com/grpc-ecosystem/grpc-gateway#installation
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# go install github.com/grpc-ecosystem/grpc-gateway/@v1
#
# go install github.com/regen-network/cosmos-proto@latest
set -eo pipefail

protoc_install_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the cosmos-sdk folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
}

protoc_install_gocosmos

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./cosmos -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# # generate codec/testdata proto code
# (cd testutil/testdata; buf generate)

# # move proto files to the right places
# cp -r github.com/cosmos/cosmos-sdk/* ./
# rm -rf github.com

# go mod tidy

# ./scripts/protocgen2.sh
