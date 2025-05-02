#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    DIR_NAME=$(echo $D | awk -F/ '{print $NF}')
    MODDOC=docs/build/modules/$DIR_NAME
    rm -rf $MODDOC
    mkdir -p $MODDOC
    if [ -f "$D/README.md" ]; then
      cp -r $D/README.md $MODDOC/
    fi
  fi
done


## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
cp ../x/auth/vesting/README.md ./docs/build/modules/auth/1-vesting.md
cp ../x/auth/tx/README.md ./docs/build/modules/auth/2-tx.md

## Add modules page list
cat ../x/README.md | sed 's/\.\.\/docs\/build\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./docs/build/modules/README.md

## Add tooling documentation
cp ../tools/cosmovisor/README.md ./docs/build/tooling/01-cosmovisor.md
cp ../tools/confix/README.md ./docs/build/tooling/02-confix.md

## Add package documentation
cp ../client/v2/README.md ./docs/learn/advanced/17-autocli.md
cp ../depinject/README.md ./docs/build/packages/01-depinject.md
cp ../collections/README.md ./docs/build/packages/02-collections.md

## Add architecture documentation
cp -r ./architecture ./docs/build

## Add spec documentation
cp -r ./spec ./docs/build

## Add rfc documentation
cp -r ./rfc ./docs/build/rfc

## Add SDK migration documentation
cp -r ../UPGRADING.md ./docs/build/migrations/02-upgrade-reference.md

cp -r ../UPGRADE_GUIDE.md ./docs/build/migrations/03-upgrade-guide.md
