#!/usr/bin/env bash

# How to run manually:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

set -eo pipefail

echo "Generating cosmos gogo proto code"
cd proto
proto_dirs=$(find ./cosmos -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

echo "Generating types/tendermint gogo proto code"
proto_dirs=$(find ./tendermint -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

echo "Generate codec/testdata proto code"
(cd testutil/testdata; buf generate)

echo "Generate baseapp test messages"
(cd baseapp/testutil; buf generate)

echo "Move proto files to the right places"
cp -r github.com/cosmos/cosmos-sdk/* ./
cp -r github.com/tendermint/tendermint/proto/tendermint/* ./types/tendermint/
cp -r github.com/tendermint/tendermint/abci* ./types/tendermint
rm -rf github.com

echo "Rewrite tendermint proto files"
find ./types/tendermint -type f -exec sed -i 's/github.com\/tendermint\/tendermint\/proto\/tendermint/github.com\/cosmos\/cosmos-sdk\/types\/tendermint/' {} \;
find ./types/tendermint -type f -exec sed -i 's/github.com\/tendermint\/tendermint\/abci\/types/github.com\/cosmos\/cosmos-sdk\/types\/abci\/types/' {} \;

go mod tidy

./scripts/protocgen-pulsar.sh
