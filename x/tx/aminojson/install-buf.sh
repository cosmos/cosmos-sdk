#!/usr/bin/env sh

BIN="$PWD/bin"
VERSION="1.9.0"

mkdir -p "$BIN"
curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
    -o "${BIN}/buf"
chmod +x "${BIN}/buf"
