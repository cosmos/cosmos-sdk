#!/usr/bin/env bash

mkdir -p modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "modules/$(echo $D | awk -F/ '{print $NF}')"
    mkdir -p "modules/$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/README.md "$_"
    if [ -f "modules/$(echo $D | awk -F/ '{print $NF}')/README.md" ]; then
      cd "modules/$(echo $D | awk -F/ '{print $NF}')"
      # This ensures that we have multiples pages for the modules documantation
      # This is easier to read for the user
      # In order to split pages, we need to add a <!-- order: X --> in the module README.md, for each pages that we want.
      csplit -k -q README.md '/<!-- order:/' '{*}' --prefix='section_' --suffix-format='%02d.md'
      rm README.md
      cd ../..
    fi
  fi
done

cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

cp ../cosmovisor/README.md ./run-node/cosmovisor.md
