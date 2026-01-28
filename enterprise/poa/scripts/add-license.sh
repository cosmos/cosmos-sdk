#!/bin/bash

set -e

# Convert license text to Go-style comments
go_license_header() {
    echo "// IMPORTANT LICENSE NOTICE"
    echo "//"
    echo "// SPDX-License-Identifier: CosmosLabs-Evaluation-Only"
    echo "//"
    echo "// This file is NOT licensed under the Apache License 2.0."
    echo "//"
    echo "// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:"
    echo "// - commercial use,"
    echo "// - production use, and"
    echo "// - redistribution."
    echo "//"
    echo "// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms."
    echo "// Copyright (c) 2026 Cosmos Labs US Inc."
    echo ""
}

# Remove all existing license headers from file
remove_license_headers() {
    local file="$1"
    local tmpfile="${file}.tmp"

    # Remove all lines from any occurrence of "IMPORTANT LICENSE NOTICE"
    # through "Copyright (c) 2026 Cosmos Labs US Inc."
    sed '/\/\/ IMPORTANT LICENSE NOTICE/,/\/\/ Copyright (c) 2026 Cosmos Labs US Inc\./d' "$file" > "$tmpfile"

    # Remove any blank lines at the start of file
    sed -i.bak '1{/^$/d;}' "$tmpfile"
    rm -f "${tmpfile}.bak"

    mv "$tmpfile" "$file"
}

# Add license to a Go or proto file
add_go_license() {
    local file="$1"

    # First remove any existing license headers (including duplicates)
    remove_license_headers "$file"

    local tmpfile="${file}.tmp"

    # Create temp file with license header at top
    go_license_header > "$tmpfile"
    cat "$file" >> "$tmpfile"
    mv "$tmpfile" "$file"
}

# Process all relevant files
process_files() {
    local pattern="$1"

    while IFS= read -r -d '' file; do
        add_go_license "$file"
    done < <(find . -type f -name "$pattern" \
        ! -path "*/build/*" \
        ! -path "*/.git/*" \
        ! -path "*/vendor/*" \
        ! -path "*/.idea/*" \
        ! -path "*/node_modules/*" \
        -print0)
}

main() {
    echo "Adding license headers to source files..."

    # Process Go files
    echo "Processing .go files..."
    process_files "*.go"

    # Process proto files
    echo "Processing .proto files..."
    process_files "*.proto"

    echo "License headers added successfully!"
}

# Allow running on specific files if provided
if [ $# -gt 0 ]; then
    for file in "$@"; do
        if [ -f "$file" ]; then
            add_go_license "$file"
        fi
    done
else
    main
fi
