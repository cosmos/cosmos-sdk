#!/usr/bin/env bash

# This script handles the synchronization of deleted files between cosmos-sdk and cosmos-sdk-docs
# It should be run after the main documentation build process

# Exit on error and print commands
set -ex

# Get the root directory of the repository
REPO_ROOT=$(git rev-parse --show-toplevel)
DOCS_DIR="$REPO_ROOT/docs"
BUILD_DIR="$DOCS_DIR/build"

# Function to check if a file exists in the source but not in the build
check_deleted_files() {
    local source_dir=$1
    local build_dir=$2
    
    # Skip if source directory doesn't exist
    if [ ! -d "$source_dir" ]; then
        echo "Source directory $source_dir does not exist, skipping..."
        return 0
    fi
    
    # Create build directory if it doesn't exist
    mkdir -p "$build_dir" || {
        echo "Failed to create build directory: $build_dir"
        return 1
    }
    
    # Find all files in the source directory
    if ! find "$source_dir" -type f | while read -r source_file; do
        # Get the relative path
        rel_path=${source_file#$source_dir/}
        build_file="$build_dir/$rel_path"
        
        # Check if the file exists in the build directory
        if [ ! -f "$build_file" ]; then
            echo "File $rel_path exists in source but not in build"
            # Remove the corresponding file in the build directory if it exists
            if [ -f "$build_file" ]; then
                rm -f "$build_file" || {
                    echo "Failed to remove file: $build_file"
                    return 1
                }
            fi
        fi
    done; then
        echo "Error processing files in $source_dir"
        return 1
    fi
    
    return 0
}

# Main execution
echo "Starting file deletion synchronization..."

# Check for deleted files in each major documentation section
check_deleted_files "$DOCS_DIR/src" "$BUILD_DIR"
check_deleted_files "$DOCS_DIR/architecture" "$BUILD_DIR/architecture"
check_deleted_files "$DOCS_DIR/spec" "$BUILD_DIR/spec"
check_deleted_files "$DOCS_DIR/rfc" "$BUILD_DIR/rfc"

# Special handling for module documentation
if [ -d "../x" ]; then
    for D in ../x/*; do
        if [ -d "${D}" ]; then
            DIR_NAME=$(echo $D | awk -F/ '{print $NF}')
            MODDOC="$BUILD_DIR/modules/$DIR_NAME"
            if [ ! -d "$MODDOC" ]; then
                echo "Module $DIR_NAME documentation directory not found in build"
                rm -rf "$MODDOC" || {
                    echo "Failed to remove module directory: $MODDOC"
                    continue
                }
            fi
        fi
    done
else
    echo "Warning: ../x directory not found, skipping module documentation check"
fi

echo "File deletion synchronization complete successfully" 
