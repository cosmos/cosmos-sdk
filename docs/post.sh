#!/usr/bin/env bash

# Exit on error and print commands
set -ex

echo "Starting post-build cleanup..."

# Run the file deletion synchronization script
if ! ./scripts/sync-deletions.sh; then
    echo "Error: File deletion synchronization failed"
    exit 1
fi

# Clean up build directory
echo "Cleaning up build directory..."

# Remove module files except _category_.json
find docs/build/modules ! -name '_category_.json' -type f -exec rm -rf {} + || {
    echo "Warning: Failed to remove some module files"
}

# Remove specific tooling and package documentation
for file in \
    docs/build/tooling/01-cosmovisor.md \
    docs/build/tooling/02-confix.md \
    docs/build/tooling/03-hubl.md \
    docs/build/packages/01-depinject.md \
    docs/build/packages/02-collections.md \
    docs/learn/advaced-concepts/17-autocli.md \
    docs/user/run-node/04-rosetta.md \
    docs/build/architecture \
    docs/build/spec \
    docs/build/rfc \
    docs/learn/advanced/17-autocli.md \
    docs/build/migrations/02-upgrading.md \
    versioned_docs \
    versioned_sidebars \
    versions.json
do
    if [ -e "$file" ]; then
        rm -rf "$file" || {
            echo "Warning: Failed to remove $file"
        }
    fi
done

echo "Post-build cleanup completed successfully"
