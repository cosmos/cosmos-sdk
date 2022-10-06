#!/usr/bin/env bash

SIDEBAR_BOILERPLATE="---\nsidebar_position: 1\n---\n\n"

# for D in ../x/*; do
#   if [ -d "${D}" ]; then
#     MODDOC=modules/$(echo $D | awk -F/ '{print $NF}')
#     rm -rf $MODDOC
#     mkdir -p $MODDOC && cp -r $D/README.md "$_/index.md"
#   fi
# done

## Vesting is a submodule of auth, but we still want to display it in docs
## TODO to be removed in https://github.com/cosmos/cosmos-sdk/issues/9958
# mkdir -p modules/vesting
# cp -r ../x/auth/vesting/README.md modules/vesting

# cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

## Add Cosmovisor documentation
echo -e "$SIDEBAR_BOILERPLATE$(cat ../cosmovisor/README.md)" > ./docs/run-node/06-cosmovisor.md

## Add depinject documentation
echo -e "$SIDEBAR_BOILERPLATE$(cat ../depinject/README.md)" > ./docs/building-apps/01-depinject.md