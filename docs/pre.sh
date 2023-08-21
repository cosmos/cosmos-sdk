#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=docs/integrate/modules/$(basename "$D")
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$MODDOC/$(basename "$D").md"
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
cp ../x/auth/vesting/README.md ./docs/integrate/modules/auth/1-vesting.md
cp ../x/auth/tx/README.md ./docs/integrate/modules/auth/2-tx.md

## Add modules page list
cat ../x/README.md | sed 's/\.\.\/docs\/integrate\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./docs/integrate/modules/"$MODDOC/$(basename "$D").md"

## Add tooling documentation
cp ../tools/proto/README.md ./docs/integrate/tooling/00-proto.md
cp ../tools/cosmovisor/README.md ./docs/integrate/tooling/01-cosmovisor.md
cp ../tools/confix/README.md ./docs/integrate/tooling/02-confix.md
cp ../tools/hubl/README.md ./docs/integrate/tooling/03-hubl.md
wget -O docs/run-node/04-rosetta.md https://raw.githubusercontent.com/cosmos/rosetta/main/README.md

## Add package documentation
cp ../client/v2/README.md ./docs/develop/advanced-concepts/17-autocli.md
cp ../depinject/README.md ./docs/integrate/packages/01-depinject.md
cp ../collections/README.md ./docs/integrate/packages/02-collections.md
cp ../orm/README.md ./docs/integrate/packages/03-orm.md

## Add architecture documentation
cp -r ./architecture ./docs/integrate

## Add spec documentation
cp -r ./spec ./docs/integrate

## Add rfc documentation
cp -r ./rfc ./docs/integrate

## Add SDK migration documentation
cp -r ../UPGRADING.md ./docs/integrate/migrations/02-upgrading.md
sed -i -e "s|./proto/README.md|../../../../proto/README.md|g" ./docs/integrate/migrations/02-upgrading.md
sed -i -e "s|./docs/docs/develop/advanced-concepts/05-encoding.md|../../develop/advanced-concepts/05-encoding.md|g" ./docs/integrate/migrations/02-upgrading.md
rm ./docs/integrate/migrations/02-upgrading.md-e