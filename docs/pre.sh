#!/usr/bin/env bash

mkdir -p modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    MODDOC=modules/$(echo $D | awk -F/ '{print $NF}')
    rm -rf $MODDOC
    mkdir -p $MODDOC && cp -r $D/README.md "$_"
    if [ -f "$MODDOC/README.md" ]; then
      cd $MODDOC
      # This ensures that we have multiples pages for the modules documantation
      # This is easier to read for the user
      # In order to split pages, we need to add a <!-- order: X --> in the module README.md, for each pages that we want.
      csplit -k -q README.md '/<!-- order:/' '{*}' --prefix='section_' --suffix-format='%02d.md'
      mv section_00.md README.md
      cd ../..
    fi
  fi
done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
mkdir -p modules/vesting
cp -r ../x/auth/vesting/README.md modules/vesting
cd modules/vesting
csplit -k -q README.md '/<!-- order:/' '{*}' --prefix='section_' --suffix-format='%02d.md'
mv section_00.md README.md
cd ../..

cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

cp ../cosmovisor/README.md ./run-node/cosmovisor.md
