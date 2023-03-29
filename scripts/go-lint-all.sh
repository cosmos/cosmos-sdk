#!/usr/bin/env bash

set -euo pipefile

export pwd=$(pwd)

for modfile in $(find . -name go.mod); do
 echo "linting $(dirname $modfile)"
 DIR=$(dirname $modfile)
 (cd $DIR; golangci-lint run ./... --fix -c $pwd/.golangci.yml)
done
