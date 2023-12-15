#!/bin/bash

# Loop over all directories in proto/cosmos/
for dir in proto/cosmos/*; do
    # Check if it's a directory
    if [ -d "$dir" ]; then
        # Extract the directory name
        dir_name=$(basename "$dir")

        if [ ! -d "x/$dir_name" ]; then
            continue
        fi

        # Create the target directory if it doesn't exist
        mkdir -p "x/$dir_name/proto/cosmos"

        # Move the directory
        mv "$dir" "x/$dir_name/proto/cosmos/"
    fi
done