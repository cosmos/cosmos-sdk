#!/usr/bin/env sh

BINARY=/simd/${BINARY:-simd}
ID=${ID:-0}
LOG=${LOG:-simd.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'simd'"
	exit 1
fi

BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"

if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

export SIMDHOME="/simd/node${ID}/simd"

if [ -d "$(dirname "${SIMDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${SIMDHOME}" "$@" | tee "${SIMDHOME}/${LOG}"
else
  "${BINARY}" --home "${SIMDHOME}" "$@"
fi
