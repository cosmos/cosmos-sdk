#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=docs/build/modules/$(basename "$D")
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$_"
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
cp ../x/auth/vesting/README.md ./docs/build/modules/auth/1-vesting.md
cp ../x/auth/tx/README.md ./docs/build/modules/auth/2-tx.md

## Add modules page list
cat ../x/README.md | sed 's/\.\.\/docs\/build\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./docs/modules/README.md

## Add tooling documentation
cp ../tools/cosmovisor/README.md ./docs/build/tooling/01-cosmovisor.md
cp ../tools/confix/README.md ./docs/build/tooling/02-confix.md
cp ../tools/hubl/README.md ./docs/build/tooling/03-hubl.md
wget -O docs/user/run-node/04-rosetta.md https://raw.githubusercontent.com/cosmos/rosetta/main/README.md

## Add package documentation
cp ../client/v2/README.md ./docs/develop/advanced/17-autocli.md
cp ../depinject/README.md ./docs/build/packages/01-depinject.md
cp ../collections/README.md ./docs/build/packages/02-collections.md
cp ../orm/README.md ./docs/build/packages/03-orm.md

## Add architecture documentation
cp -r ./architecture ./docs/build

## Add spec documentation
cp -r ./spec ./docs/build

## Add rfc documentation
cp -r ./rfc ./docs/build

## Add SDK migration documentation
cp -r ../UPGRADING.md ./docs/build/migrations/02-upgrading.md