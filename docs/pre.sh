#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=docs/modules/$(echo $D | awk -F/ '{print $NF}')
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$_"
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
mkdir -p docs/modules/vesting
cp ../x/auth/vesting/README.md ./docs/modules/vesting/README.md

## Add modules page list
cat ../x/README.md | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./docs/modules/README.md

### Notable packages documentation

cp ../x/auth/tx/README.md ./docs/pkg/01-authtx.md
cp ../depinject/README.md ./docs/pkg/02-depinject.md
cp ../snapshots/README.md ./docs/pkg/03-snapshots.md

mkdir -p docs/pkg/store
cp ../store/pruning/README.md ./docs/pkg/store/01-pruning.md
cp ../store/streaming/README.md ./docs/pkg/store/02-streaming.md

### Tooling

## Add cosmovisor documentation
cp ../tools/cosmovisor/README.md ./docs/tooling/01-cosmovisor.md

## Add rosetta documentation
cp ../tools/rosetta/README.md ./docs/run-node/04-rosetta.md

### Others

## Add architecture documentation
cp -r ./architecture ./docs

## Add spec documentation
cp -r ./spec ./docs
