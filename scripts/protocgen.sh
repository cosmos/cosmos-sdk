#!/usr/bin/env bash

set -eo pipefail

protoc_install_gocosmos() {
   echo "Installing protobuf gocosmos plugin"
   # we should use go install, but regen-network/cosmos-proto contains
   # replace directives. It must not contain directives that would cause
   # it to be interpreted differently than if it were the main module.
   # So the command below issues a warning and we are muting it for now.
   #
   # Installing plugins must be done outside of the module
   (go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@v0.3.1 2> /dev/null)
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

# generate codec/testdata proto code
(cd testutil/testdata; buf generate)

# move proto files to the right places
cp -r github.com/cosmos/cosmos-sdk/* ./
rm -rf github.com

go mod tidy

./scripts/protocgen2.sh
