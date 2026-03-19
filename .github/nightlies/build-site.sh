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

if [ "$IS_MAIN" = "true" ]; then
  cat > "$NIGHTLIES_DIR/latest.json" <<EOF
{
  "channel": "latest",
  "redirect": "./latest/latest.json"
}
EOF
fi

CHANNELS_JSON="$NIGHTLIES_DIR/channels.json"

channels_array="[]"

for dir in "$NIGHTLIES_DIR"/*/; do
  [ -d "$dir" ] || continue
  name=$(basename "$dir")

  channels_array=$(echo "$channels_array" | jq \
    --arg n "$name" \
    '. += [{
      "name": $n,
      "path": $n,
      "manifest": "./\($n)/latest.json"
    }]')
done

echo "{ \"channels\": $channels_array }" | jq . > "$CHANNELS_JSON"

cat > "$NIGHTLIES_DIR/index.html" <<EOF
<meta http-equiv="refresh" content="0; url=./latest/">
EOF
