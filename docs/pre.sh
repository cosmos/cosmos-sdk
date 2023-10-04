#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=build/modules/$(echo $D | awk -F/ '{print $NF}')
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$_"
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
cp -r ../x/auth/vesting/README.md ./build/modules/auth/1-vesting.md
cp -r ../x/auth/tx/README.md ./build/modules/auth/2-tx.md

## Add modules page list
cat ../x/README.md | sed 's/\.\.\/\/build\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./build/modules/README.md

## Add tooling documentation
cp ../tools/cosmovisor/README.md ./build/tooling/01-cosmovisor.md
cp ../tools/confix/README.md ./build/tooling/02-confix.md
cp ../tools/hubl/README.md ./build/tooling/03-hubl.md

wget -x -O ./user/run-node/04-rosetta.md https://raw.githubusercontent.com/cosmos/rosetta/main/README.md

## Add package documentation
cp ../client/v2/README.md ./learn/advanced/17-autocli.md
cp ../depinject/README.md ./build/packages/01-depinject.md
cp ../collections/README.md ./build/packages/02-collections.md
cp ../orm/README.md ./build/packages/03-orm.md

## Add architecture documentation
cp -r ./architecture ./build

## Add spec documentation
cp -r ./spec ./build

## Add rfc documentation
cp -r ./rfc ./build

## Add SDK migration documentation
cp -r ../UPGRADING.md ./build/migrations/02-upgrading.md
