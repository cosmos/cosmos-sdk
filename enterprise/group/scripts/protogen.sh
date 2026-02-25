#!/bin/sh

set -ex

cd proto
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..
# Copy gogo-generated .pb.go and .pb.gw.go files to x/group
cp ./github.com/cosmos/cosmos-sdk/enterprise/group/*.pb.go ./x/group/ 2>/dev/null || true
cp ./github.com/cosmos/cosmos-sdk/enterprise/group/*.pb.gw.go ./x/group/ 2>/dev/null || true

rm -rf ./github.com ./cosmos

# Add license headers to all generated files
echo "Adding license headers to generated files..."
sh ./scripts/add-license.sh
