#!/usr/bin/env bash

set -euo pipefail

for modfile in $(find . -name go.mod); do
 echo "Updating $modfile"
 DIR=$(dirname $modfile)
 (cd $DIR; go mod tidy)
done

# update gomod2nix.toml for simapp
# NOTE: gomod2nix should be built using the same go version as the project, the nix flake will make sure of that
# automatically.
cd simapp
if command -v nix &> /dev/null
    nix develop .. -c gomod2nix
then
    if ! command -v gomod2nix &> /dev/null
    then
        echo "gomod2nix could not be found in PATH, installing..."
        go install github.com/nix-community/gomod2nix@latest
    fi
    gomod2nix
fi
