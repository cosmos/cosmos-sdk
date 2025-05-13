#!/usr/bin/env bash

# How to run manually:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

echo "Formatting protobuf files"
find ./ -name "*.proto" -exec clang-format -i {} \;

set -e

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./cosmos ./amino -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    # this regex checks if a proto file has its go_package set to cosmossdk.io/api/...
    # gogo proto files SHOULD ONLY be generated if this is false
    # we don't want gogo proto to run for proto files which are natively built for google.golang.org/protobuf
    if grep -q "option go_package" "$file" && grep -H -o -c 'option go_package.*cosmossdk.io/api' "$file" | grep -q ':0$'; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# generate tests proto code
echo "Generating tests proto code"
(cd testutil/testdata; buf generate)
(cd baseapp/testutil; buf generate)
(cd tests/integration/tx/internal; make codegen)

# move proto files to the right places
echo "Moving proto files"
cp -r github.com/cosmos/cosmos-sdk/* ./
cp -r cosmossdk.io/** ./
rm -rf github.com cosmossdk.io

echo "Generating pulsar proto code"
./scripts/protocgen-pulsar.sh

echo
echo "All Protobuf code generation steps completed"
echo "Last step: running go mod tidy"
echo
go mod tidy
