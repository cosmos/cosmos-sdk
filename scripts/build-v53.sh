#!/usr/bin/env bash
# build-v53 fetches the v0.53 simd binary for system tests.
# Downloads from the v0.53.x-nightly release (non-production).
# If download fails, errors with instructions to build manually.
set -euo pipefail

REPO="cosmos/cosmos-sdk"
RELEASE_TAG="v0.53.x-nightly"
OUTPUT="${BUILDDIR:-./build}/simdv53"

# Map uname to Go-style GOOS/GOARCH
detect_goos_goarch() {
	local u
	u=$(uname -s)
	local m
	m=$(uname -m)

	case "$u" in
		Linux)  GOOS=linux ;;
		Darwin) GOOS=darwin ;;
		*)
			echo "Unsupported OS: $u" >&2
			return 1
			;;
	esac

	case "$m" in
		x86_64|amd64)  GOARCH=amd64 ;;
		arm64|aarch64) GOARCH=arm64 ;;
		*)
			echo "Unsupported arch: $m" >&2
			return 1
			;;
	esac

	echo "${GOOS}-${GOARCH}"
}

try_download() {
	local goos_goarch
	goos_goarch=$(detect_goos_goarch)
	local asset="simd-${goos_goarch}"
	local url="https://github.com/${REPO}/releases/download/${RELEASE_TAG}/${asset}"

	echo "Attempting to download ${asset} from ${RELEASE_TAG}..."
	if curl -sfL -o "$OUTPUT" "$url"; then
		chmod +x "$OUTPUT"
		echo "Downloaded simdv53 to ${OUTPUT}"
		return 0
	fi
	return 1
}

main() {
	mkdir -p "$(dirname "$OUTPUT")"
	if try_download; then
		exit 0
	fi

	local root_dir
	root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
	echo "" >&2
	echo "Download failed. To build manually:" >&2
	echo "  cd ${root_dir}" >&2
	echo "  git checkout release/v0.53.x" >&2
	echo "  make build" >&2
	echo "  cp build/simd \"${OUTPUT}\"" >&2
	echo "" >&2
	exit 1
}

main "$@"
