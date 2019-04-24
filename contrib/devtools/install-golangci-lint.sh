#!/bin/bash

set -euo pipefail

installer="$(mktemp)"
trap "rm -f ${installer}" EXIT

GOBIN="${1}"
VERSION="${2}"
HASHSUM="${3}"
CURL="$(which curl)"
GOSUM="$(which gosum)"

echo "Downloading golangci-lint ${VERSION} installer ..." >&2
"${CURL}" -sfL "https://raw.githubusercontent.com/golangci/golangci-lint/${VERSION}/install.sh" > "${installer}"

echo "Checking hashsum ..." >&2
[ "${HASHSUM}" = "$(${GOSUM} ${installer})" ]
chmod +x "${installer}"

echo "Launching installer ..." >&2
exec "${installer}" -d -b "${GOBIN}" "${VERSION}"
