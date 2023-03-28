#!/usr/bin/env bash

set -euo pipefail

export pwd=$(pwd)

for modfile in $(find . -name go.mod); do
 echo "Updating $modfile"
 DIR=$(dirname $modfile)
 (cd $DIR; golangci-lint run ./... --fix -c $pwd/.golangci.yml)
done
