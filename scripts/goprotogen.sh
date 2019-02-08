#!/bin/bash

set -e

PROTO_OUT=$1
MODULES=$(go list ./x/...)
MODULES="$MODULES github.com/cosmos/cosmos-sdk/types"

if [ -z "$PROTO_OUT" ]
then
    echo "proto source directory is empty; exiting"
    exit 1
fi

mkdir -p $PROTO_OUT

echo "generating proto files..."
for mod in ${MODULES}; do
    set -o xtrace
    proteus proto -f $PROTO_OUT -p $mod
    { set +o xtrace; } 2>/dev/null
done
