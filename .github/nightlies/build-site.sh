#!/usr/bin/env bash
set -euo pipefail

CHANNEL="$1"
IS_MAIN="$2"

OWNER="${GITHUB_REPOSITORY_OWNER}"
REPO="${GITHUB_REPOSITORY##*/}"
SHA="${GITHUB_SHA}"

SITE_DIR="site"
NIGHTLIES_DIR="$SITE_DIR/nightlies"
CHANNEL_DIR="$NIGHTLIES_DIR/$CHANNEL"

mkdir -p "$SITE_DIR"
cp -r existing/* "$SITE_DIR" 2>/dev/null || true

mkdir -p "$CHANNEL_DIR"

cp artifacts/*.tar.gz "$CHANNEL_DIR/"
cp artifacts/*.sha256 "$CHANNEL_DIR/"

BASE_URL="https://${OWNER}.github.io/${REPO}/nightlies/${CHANNEL}"

# channel manifest
cat > "$CHANNEL_DIR/latest.json" <<EOF
{
  "channel": "${CHANNEL}",
  "commit": "${SHA}",
  "base_url": "${BASE_URL}",
  "assets": {
    "linux-amd64": {
      "url": "${BASE_URL}/simd-linux-amd64.tar.gz",
      "sha256_url": "${BASE_URL}/simd-linux-amd64.tar.gz.sha256"
    },
    "linux-arm64": {
      "url": "${BASE_URL}/simd-linux-arm64.tar.gz",
      "sha256_url": "${BASE_URL}/simd-linux-arm64.tar.gz.sha256"
    },
    "darwin-amd64": {
      "url": "${BASE_URL}/simd-darwin-amd64.tar.gz",
      "sha256_url": "${BASE_URL}/simd-darwin-amd64.tar.gz.sha256"
    },
    "darwin-arm64": {
      "url": "${BASE_URL}/simd-darwin-arm64.tar.gz",
      "sha256_url": "${BASE_URL}/simd-darwin-arm64.tar.gz.sha256"
    }
  }
}
EOF

mkdir -p "$NIGHTLIES_DIR"

# global latest.json (only main updates)
if [ "$IS_MAIN" = "true" ]; then
  cat > "$NIGHTLIES_DIR/latest.json" <<EOF
{
  "channel": "latest",
  "redirect": "./latest/latest.json"
}
EOF
fi

# channels.json
CHANNELS_JSON="$NIGHTLIES_DIR/channels.json"
echo '{ "channels": [' > "$CHANNELS_JSON"

first=true
for dir in "$NIGHTLIES_DIR"/*/; do
  [ -d "$dir" ] || continue
  name=$(basename "$dir")

  if [ "$first" = true ]; then
    first=false
  else
    echo ',' >> "$CHANNELS_JSON"
  fi

  cat >> "$CHANNELS_JSON" <<EOF
{
  "name": "$name",
  "path": "$name",
  "manifest": "./$name/latest.json"
}
EOF
done

echo ']}' >> "$CHANNELS_JSON"

# redirect
cat > "$NIGHTLIES_DIR/index.html" <<EOF
<meta http-equiv="refresh" content="0; url=./latest/">
EOF