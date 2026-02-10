#!/bin/sh

set -ex

cd proto
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..
cp -r ./github.com/cosmos/cosmos-sdk/enterprise/poa/types/* ./x/poa/types

rm -rf ./github.com

# Add license headers to all generated files
echo "Adding license headers to generated files..."
sh ./scripts/add-license.sh