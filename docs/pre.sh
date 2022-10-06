#!/usr/bin/env bash

mkdir -p modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=modules/$(echo $D | awk -F/ '{print $NF}')
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$_"
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
mkdir -p modules/vesting
cp -r ../x/auth/vesting/README.md modules/vesting

cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

## Add Cosmovisor documentation
cp ../cosmovisor/README.md ./run-node/cosmovisor.md

## Add depinject documentation
cp ../depinject/README.md ./building-apps/depinject.md