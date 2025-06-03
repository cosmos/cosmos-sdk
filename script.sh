#!/bin/sh

# macOS-compatible relpath using Python
relpath() {
  python3 -c "import os.path; print(os.path.relpath('$1', '$2'))"
}

ROOT_DIR=$(pwd)
COMET_DIR="$ROOT_DIR/../cometbft"
COMET_API_DIR="$ROOT_DIR/../cometbft/api"

find . -name 'go.mod' | while read -r modfile; do
    MOD_DIR=$(cd "$(dirname "$modfile")" && pwd)
    cd "$MOD_DIR" || exit 1

    REL_COMET_DIR=$(relpath "$COMET_DIR" "$MOD_DIR")
    REL_COMET_API_DIR=$(relpath "$COMET_API_DIR" "$MOD_DIR")

    go mod edit -replace=github.com/cometbft/cometbft="$REL_COMET_DIR"
    go mod edit -replace=github.com/cometbft/cometbft/api="$REL_COMET_API_DIR"

    echo "Updated $modfile in $MOD_DIR"

    if ! go mod tidy; then
        echo "go mod tidy failed in $MOD_DIR"
        exit 1
    fi

    cd "$ROOT_DIR" || exit 1
done

