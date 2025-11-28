#!/usr/bin/env bash

set -e

echo "Generating gogo proto code"
proto_dirs=$(find ./contrib -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    # this regex checks if a proto file has its go_package set to cosmossdk.io/api/...
    # gogo proto files SHOULD ONLY be generated if this is false
    # we don't want gogo proto to run for proto files which are natively built for google.golang.org/protobuf
    if grep -q "option go_package" "$file" && grep -H -o -c 'option go_package.*cosmossdk.io/api' "$file" | grep -q ':0$'; then
      echo buf generate --template buf.gen.gogo.yaml $file
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cp -r ../github.com/cosmos/cosmos-sdk/contrib/* ../
rm -rf ../github.com

buf generate --template buf.gen.pulsar.yaml
