#!/usr/bin/env bash

## Create modules pages
for D in ../x/*; do
  if [ -d "${D}" ]; then
    DIR_NAME=$(basename "$D")
    
    # Skip specific directories
    if [[ "$DIR_NAME" != "counter" ]]; then
      MODULE_DIRECTORY=docs/build/modules/$DIR_NAME
      rm -rf "$MODULE_DIRECTORY"
      mkdir -p "$MODULE_DIRECTORY" 
      if [ -f "$D"/README.md ]; then
        cp -r "$D"/README.md "$MODULE_DIRECTORY"
      fi
    fi
  fi
done

cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/spec\/README.md//g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

cp ../cosmovisor/README.md ./run-node/cosmovisor.md
