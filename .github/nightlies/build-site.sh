#!/usr/bin/env bash
set -euo pipefail

CHANNEL="$1"
IS_MAIN="$2"

OWNER="${GITHUB_REPOSITORY_OWNER}"
REPO="${GITHUB_REPOSITORY##*/}"
SHA="${GITHUB_SHA}"
SHORT_SHA="${SHA:0:7}"

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

# HTML UI
cat > "$CHANNEL_DIR/index.html" <<EOF
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>simd ${CHANNEL} nightlies</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, sans-serif;
      max-width: 700px;
      margin: 40px auto;
      padding: 0 16px;
      background: #fafafa;
      color: #111;
    }
    h1 {
      margin-bottom: 4px;
    }
    .meta {
      color: #666;
      margin-bottom: 24px;
      font-size: 14px;
    }
    ul {
      list-style: none;
      padding: 0;
    }
    li {
      margin: 8px 0;
    }
    a {
      text-decoration: none;
      color: #0969da;
    }
    a:hover {
      text-decoration: underline;
    }
    .box {
      background: white;
      border: 1px solid #e5e5e5;
      border-radius: 8px;
      padding: 16px;
    }
  </style>
</head>
<body>
  <h1>simd ${CHANNEL} nightlies</h1>
  <div class="meta">
    commit <a href="https://github.com/${OWNER}/${REPO}/commit/${SHA}">
      <code>${SHORT_SHA}</code>
    </a>
  </div>

  <div class="box">
    <ul>
      <li><a href="./simd-linux-amd64.tar.gz">Linux (amd64)</a> - <a href="./simd-linux-amd64.tar.gz.sha256">sha256</a></li>
      <li><a href="./simd-linux-arm64.tar.gz">Linux (arm64)</a> - <a href="./simd-linux-arm64.tar.gz.sha256">sha256</a></li>
      <li><a href="./simd-darwin-amd64.tar.gz">macOS (amd64)</a> - <a href="./simd-darwin-amd64.tar.gz.sha256">sha256</a></li>
      <li><a href="./simd-darwin-arm64.tar.gz">macOS (arm64)</a> - <a href="./simd-darwin-arm64.tar.gz.sha256">sha256</a></li>
    </ul>
  </div>

  <p style="margin-top:20px;">
    <a href="./latest.json">latest.json</a>
  </p>
</body>
</html>
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
