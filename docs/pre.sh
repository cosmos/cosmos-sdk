#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    DIR_NAME=$(basename "$D")

    # Skip specific directories
    if [[ "$DIR_NAME" != "counter" ]]; then
      MODULE_DIRECTORY=./docs/build/modules/$DIR_NAME
      rm -rf "$MODULE_DIRECTORY"
      mkdir -p "$MODULE_DIRECTORY"
      if [ -f "$D"/README.md ]; then
        cp -r "$D"/README.md "$MODULE_DIRECTORY"
      fi
    fi
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
cp -r ../x/auth/vesting/README.md ./build/modules/auth/1-vesting.md
cp -r ../x/auth/tx/README.md ./build/modules/auth/2-tx.md

## Add modules page list
cat ../x/README.md | sed 's#../build/modules/README.md#/building-modules/intro.html#g' > ./docs/build/modules/README.md

## Add tooling documentation
cp ../tools/cosmovisor/README.md ./docs/build/tooling/01-cosmovisor.md
cp ../tools/confix/README.md ./docs/build/tooling/02-confix.md
cp ../tools/hubl/README.md ./docs/build/tooling/03-hubl.md

## Add package documentation
cp ../client/v2/README.md ./docs/learn/advanced/17-autocli.md
cp ../depinject/README.md ./docs/build/packages/01-depinject.md
cp ../collections/README.md ./docs/build/packages/02-collections.md
cp ../orm/README.md ./docs/build/packages/03-orm.md

## Update user docs with rosetta
wget -O "./docs/user/run-node/04-rosetta.md" "https://raw.githubusercontent.com/cosmos/rosetta/main/README.md"

# TODO: tbh this should be in the docs/ directly. This way each future tagged version gets its version of the arch, spec, etc
## Add architecture documentation
# cp -r ./architecture ./build

## Add spec documentation
# cp -r ./spec ./build

## Add rfc documentation
# cp -r ./rfc ./build

## Add SDK migration documentation
cp -r ../UPGRADING.md ./docs/build/migrations/02-upgrading.md
