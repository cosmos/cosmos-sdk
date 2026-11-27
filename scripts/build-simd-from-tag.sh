#!/usr/bin/env bash
# build-simd-from-tag builds simd from a released cosmos-sdk tag for the
# upgrade system test.
#
# Usage: build-simd-from-tag.sh <tag> <output>
#
# The binary is built with the Go version in that tag's go.mod, because older
# releases do not build with newer Go toolchains (for example, bytedance/sonic
# in v0.53.x fails on Go 1.26).
set -euo pipefail

if [ "$#" -ne 2 ]; then
	echo "usage: $0 <tag> <output>" >&2
	exit 1
fi

TAG="$1"
OUTPUT="$2"
REPO_URL="${COSMOS_SDK_REPO_URL:-https://github.com/cosmos/cosmos-sdk.git}"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

echo "Building simd ${TAG}..."
git clone --quiet --depth 1 --branch "${TAG}" "${REPO_URL}" "${tmp_dir}/cosmos-sdk"

go_version="$(GOTOOLCHAIN=local go mod edit -json "${tmp_dir}/cosmos-sdk/go.mod" | sed -n 's/^[[:space:]]*"Go": "\(.*\)",$/\1/p')"
if [ -z "${go_version}" ]; then
	echo "could not read the Go version from ${TAG} go.mod" >&2
	exit 1
fi

(cd "${tmp_dir}/cosmos-sdk" && GOTOOLCHAIN="go${go_version}" make build BUILDDIR="${tmp_dir}/build")

mkdir -p "$(dirname "${OUTPUT}")"
mv "${tmp_dir}/build/simd" "${OUTPUT}"
chmod +x "${OUTPUT}"
echo "Built simd ${TAG} with go${go_version} to ${OUTPUT}"
