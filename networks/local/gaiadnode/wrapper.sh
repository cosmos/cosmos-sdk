#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/gaiad/${BINARY:-gaiad}
ID=${ID:-0}
LOG=${LOG:-gaiad.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'gaiad' E.g.: -e BINARY=gaiad_my_test_version"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export GAIADHOME="/gaiad/node${ID}/gaiad"

if [ -d "`dirname ${GAIADHOME}/${LOG}`" ]; then
  "$BINARY" --home "$GAIADHOME" "$@" | tee "${GAIADHOME}/${LOG}"
else
  "$BINARY" --home "$GAIADHOME" "$@"
fi

chmod 777 -R /gaiad

