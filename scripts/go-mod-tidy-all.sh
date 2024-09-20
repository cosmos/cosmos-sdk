#!/usr/bin/env bash

set -euo pipefail

for modfile in $(find . -name go.mod); do
 echo "Updating $modfile"
 DIR=$(dirname $modfile)
 (cd $DIR; go mod tidy -go=1.23)
done
