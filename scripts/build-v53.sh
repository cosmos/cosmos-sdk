#!/usr/bin/env bash
# build-v53 fetches the v0.53 simd binary for system tests.
# Downloads from the v0.53 nightly channel on GitHub Pages (non-production).
# If download fails, errors with instructions to build manually.
set -euo pipefail

NIGHTLY_BASE_URL="https://cosmos.github.io/cosmos-sdk/nightlies/v0.53"
OUTPUT="${BUILDDIR:-./build}/simdv53"

sha256_file() {
	local file="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
		return 0
	fi
	shasum -a 256 "$file" | awk '{print $1}'
}

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
	local archive="${asset}.tar.gz"
	local archive_url="${NIGHTLY_BASE_URL}/${archive}"
	local checksum_url="${archive_url}.sha256"
	local tmp_dir
	tmp_dir="$(mktemp -d)"
	local archive_path="${tmp_dir}/${archive}"
	local checksum_path="${tmp_dir}/${archive}.sha256"

	cleanup() {
		rm -rf "${tmp_dir}"
	}
	trap cleanup RETURN

	echo "Attempting to download ${archive} from v0.53 nightlies..."
	curl -sfL -o "${archive_path}" "${archive_url}"
	curl -sfL -o "${checksum_path}" "${checksum_url}"

	local expected actual
	expected="$(awk '{print $1}' "${checksum_path}")"
	actual="$(sha256_file "${archive_path}")"
	if [ "${expected}" != "${actual}" ]; then
		echo "Checksum mismatch for ${archive}" >&2
		return 1
	fi

	tar -xzf "${archive_path}" -C "${tmp_dir}"
	mv "${tmp_dir}/${asset}" "$OUTPUT"
	chmod +x "$OUTPUT"
	echo "Downloaded simdv53 to ${OUTPUT}"
	return 0
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
