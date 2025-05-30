#!/bin/sh

ROOT_DIR=$(pwd)
TARGET_COMMIT="v0.0.0-20250530022340-7109c1e6ee31" # pseudo-version

find . -name 'go.mod' | while read -r modfile; do
    MOD_DIR=$(cd "$(dirname "$modfile")" && pwd)
    cd "$MOD_DIR" || exit 1

    go mod edit -replace=github.com/cometbft/cometbft=github.com/zrbecker/cometbft@$TARGET_COMMIT
    go mod edit -replace=github.com/cometbft/cometbft/api=github.com/zrbecker/cometbft/api@$TARGET_COMMIT

    echo "Updated $modfile in $MOD_DIR"

    if ! go mod tidy; then
        echo "go mod tidy failed in $MOD_DIR"
        exit 1
    fi

    cd "$ROOT_DIR" || exit 1
done
