#!/bin/bash

set -e

# Usage: add-license.sh <module_dir> <license_path>
#   module_dir:  Full path to the enterprise module (e.g. /path/to/enterprise/group)
#   license_path: Path for LICENSE link in header (e.g. enterprise/group)
#
# Must be run from repo root or with absolute paths.

MODULE_DIR="$1"
LICENSE_PATH="$2"

if [[ -z "$MODULE_DIR" || -z "$LICENSE_PATH" ]]; then
    echo "Usage: add-license.sh <module_dir> <license_path>"
    echo "  e.g. add-license.sh /path/to/enterprise/group enterprise/group"
    exit 1
fi

if [[ ! -d "$MODULE_DIR" ]]; then
    echo "Error: Module directory does not exist: $MODULE_DIR"
    exit 1
fi

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
    echo "// See https://github.com/cosmos/cosmos-sdk/blob/main/${LICENSE_PATH}/LICENSE for full terms."
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
    done < <(find "$MODULE_DIR" -type f -name "$pattern" \
        ! -path "*/build/*" \
        ! -path "*/.git/*" \
        ! -path "*/vendor/*" \
        ! -path "*/.idea/*" \
        ! -path "*/node_modules/*" \
        -print0)
}

main() {
    echo "Adding license headers to source files in $MODULE_DIR..."

    # Process Go files
    echo "Processing .go files..."
    process_files "*.go"

    # Process proto files
    echo "Processing .proto files..."
    process_files "*.proto"

    echo "License headers added successfully!"
}

# Allow running on specific files if provided
if [ $# -gt 2 ]; then
    shift 2
    for file in "$@"; do
        if [ -f "$file" ]; then
            add_go_license "$file"
        fi
    done
else
    main
fi
