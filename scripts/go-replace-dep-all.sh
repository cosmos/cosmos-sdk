#!/usr/bin/env bash

set -euo pipefail

if [ $# -lt 2 ]; then
  echo "USAGE:
    ./scripts/go-replace-dep-all.sh <go-mod-dependency> <go-mod-replacement>
  This command replaces a dependency in all of the go.mod files which import it.
  It should be called with two arguments which is the go module path of the
  dependency, and the go module path of the dependency that should replace it."
  exit 1
fi

dependency=$1
replacement=$2
# in case the user explicitly specified a dependency version with @, we separate
# the dependency module name into dependency_mod
IFS='@' read -ra dependency_mod <<< "$dependency"
dependency_mod=${dependency_mod[0]}

while IFS= read -r modfile; do
  if grep -q "$dependency_mod" "$modfile"; then
    echo "Updating $modfile"
    DIR=$(dirname "$modfile")
    # we want to skip the go.mod of the package we're updating
    if [[ "$dependency_mod" == *"$(basename "$DIR")" ]]; then
        echo "Skipping $DIR"
        continue
    fi
    (cd "$DIR"; go mod edit -replace "$dependency=$replacement")
  fi
done < <(find . -name go.mod)
