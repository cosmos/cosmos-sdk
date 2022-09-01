#!/usr/bin/env bash

SIMAPP_REGEX="cosmossdk.io/simapp|github.com/cosmos/cosmos-sdk/simapp"

for d in $(find . -name 'go.mod' | xargs -L 1 dirname)
do
    if $(cd $d && go list -f '{{ .Imports }}' ./... | grep -E '${SIMAPP_REGEX}'); then
        echo "${d} has a dependency on simapp!"
        exit 1
    fi
done
