#!/usr/bin/env bash
# build-v54 fetches the v0.54 simd binary for system tests.
# Downloads from the v0.54 nightly channel on GitHub Pages (non-production).
# If download fails, errors with instructions to build manually.
set -euo pipefail

NIGHTLY_BASE_URL="https://cosmos.github.io/cosmos-sdk/nightlies/v0.54"
OUTPUT="${BUILDDIR:-./build}/simdv54"

ATTESTATION_REPO="${NIGHTLY_ATTESTATION_REPO:-cosmos/cosmos-sdk}"
ATTESTATION_WORKFLOW="${ATTESTATION_REPO}/.github/workflows/build-simd-nightlies.yml"

verify_attestation() {
	local file="$1"

	if ! command -v gh >/dev/null 2>&1; then
		echo "gh CLI is required to verify nightly build provenance." >&2
		echo "Install it from https://cli.github.com/ (and 'gh auth login')." >&2
		return 1
	fi

	if ! gh attestation verify "$file" \
		--repo "$ATTESTATION_REPO" \
		--signer-workflow "$ATTESTATION_WORKFLOW" >/dev/null 2>&1; then
		echo "Provenance verification failed for ${file}." >&2
		echo "The archive was not attested by ${ATTESTATION_WORKFLOW}." >&2
		return 1
	fi
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
	local tmp_dir
	tmp_dir="$(mktemp -d)"
	local archive_path="${tmp_dir}/${archive}"

	cleanup() {
		rm -rf "${tmp_dir}"
	}
	trap cleanup RETURN

	echo "Attempting to download ${archive} from v0.54 nightlies..."
	if ! curl -sfL -o "${archive_path}" "${archive_url}"; then
		echo "Failed to download ${archive} from ${archive_url}" >&2
		return 1
	fi

	echo "Verifying build provenance for ${archive}..."
	if ! verify_attestation "${archive_path}"; then
		return 1
	fi

	tar -xzf "${archive_path}" -C "${tmp_dir}"
	mv "${tmp_dir}/${asset}" "$OUTPUT"
	chmod +x "$OUTPUT"
	echo "Downloaded simdv54 to ${OUTPUT}"
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
	echo "  git checkout release/v0.54.x" >&2
	echo "  make build" >&2
	echo "  cp build/simd \"${OUTPUT}\"" >&2
	echo "" >&2
	exit 1
}

main "$@"
