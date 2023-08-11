#!/usr/bin/env bash

set -euo pipefail

for modfile in $(find . -name go.mod); do
 echo "Updating $modfile"
 DIR=$(dirname $modfile)
 (cd $DIR; go mod tidy)
done

if ! command -v gomod2nix &> /dev/null
then
    echo "gomod2nix could not be found in PATH, installing..."
    go install github.com/nix-community/gomod2nix@latest
fi
# update gomod2nix.toml for simapp
cd simapp; gomod2nix
