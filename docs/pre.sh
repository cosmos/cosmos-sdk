#!/usr/bin/env bash

## Create modules pages
mkdir -p docs/modules
cp modules_category.json docs/modules/_category_.json

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
cp -r ../x/auth/vesting/README.md ./docs/modules/vesting/README.md

## Add modules page list
cat ../x/README.md | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./docs/modules/README.md

## Add Cosmovisor documentation
cp ../cosmovisor/README.md ./docs/run-node/06-cosmovisor.md

## Add depinject documentation
cp ../depinject/README.md ./docs/building-apps/01-depinject.md

## Add architecture documentation
cp -r ./architecture ./docs

## Add spec documentation
cp -r ./spec ./docs