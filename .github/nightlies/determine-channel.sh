#!/usr/bin/env bash
set -euo pipefail

REF="${GITHUB_REF_NAME}"

if [ "$REF" = "main" ]; then
  echo "channel=latest" >> "$GITHUB_OUTPUT"
  echo "is_main=true" >> "$GITHUB_OUTPUT"

elif [[ "$REF" =~ release/v([0-9]+\.[0-9]+)\.x ]]; then
  echo "channel=v${BASH_REMATCH[1]}" >> "$GITHUB_OUTPUT"
  echo "is_main=false" >> "$GITHUB_OUTPUT"

else
  echo "Unsupported branch: $REF. Expected main or release/vX.Y.x."
  exit 1
fi
