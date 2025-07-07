#!/usr/bin/env bash

set -euo pipefail

# Deal with the root
go mod tidy -go=1.24

# Find all go.mod files, excluding the one in the root directory
for modfile in $(find . -mindepth 2 -name go.mod); do
  echo "Updating $modfile"
  DIR=$(dirname $modfile)
  (cd $DIR; go mod tidy -go=1.24 && go get github.com/cosmos/cosmos-sdk@v0.50.13 && go get github.com/cometbft/cometbft@v0.38.18)
done
