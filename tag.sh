#! /bin/bash
set -ex
suffix=-alpha.agoric.0
VERSION=v0.53.6$suffix

xs="api collections core depinject errors log math simapp store systemtests tests tests/systemtests tools/benchmark tools/confix tools/cosmovisor x/circuit x/evidence x/feegrant x/nft x/tx x/upgrade"
tags=

for x in "$VERSION" "client/v2.0.0-beta.11-$VERSION"; do
  git tag -f "$x"
  tags="$tags $x"
done
for x in $xs; do
  git tag -f "$x/$VERSION"
  tags="$tags $x/$VERSION"
done
git push -fu origin$tags
