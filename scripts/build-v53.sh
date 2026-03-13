#!/usr/bin/env bash
# build-v53 fetches the v0.53 simd binary for system tests.
# Tries to download from the v0.53.x-nightly release (non-production) first.
# Falls back to building from source if download fails.
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
			echo "Unsupported OS: $u"
			return 1
			;;
	esac

	case "$m" in
		x86_64|amd64)  GOARCH=amd64 ;;
		arm64|aarch64) GOARCH=arm64 ;;
		*)
			echo "Unsupported arch: $m"
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

build_from_source() {
	echo "Download failed, building from source..."
	local script_dir
	script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
	local root_dir
	root_dir="$(cd "${script_dir}/.." && pwd)"

	(
		cd "$root_dir"
		git_status=$(git status --porcelain)
		has_changes=false
		if [ -n "$git_status" ]; then
			echo "Stashing uncommitted changes..."
			git stash push -m "Temporary stash for v53 build" || true
			has_changes=true
		fi

		CURRENT_REF=$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse HEAD)
		echo "Checking out release branch..."
		git checkout release/v0.53.x
		echo "Building v53 binary..."
		make build
		mkdir -p "$(dirname "$OUTPUT")"
		mv build/simd "$OUTPUT"
		echo "Returning to original branch..."
		if [ "$CURRENT_REF" = "HEAD" ]; then
			git checkout "$(git rev-parse HEAD)"
		else
			git checkout "$CURRENT_REF"
		fi

		if [ "$has_changes" = "true" ]; then
			echo "Reapplying stashed changes..."
			git stash pop || echo "Warning: Could not pop stash, your changes may be in the stash list"
		fi
	)

	echo "Built simdv53 at ${OUTPUT}"
}

main() {
	mkdir -p "$(dirname "$OUTPUT")"
	if try_download; then
		exit 0
	fi
	build_from_source
}

main "$@"
