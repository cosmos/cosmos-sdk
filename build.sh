#!/bin/bash

set -ue

# Expect the following envvars to be set:
# - APP
# - VERSION
# - COMMIT
# - TARGET_OS
# - LEDGER_ENABLED
# - DEBUG

# Source builder's functions library
. /usr/local/share/cosmos-sdk/buildlib.sh

# These variables are now available
# - BASEDIR
# - OUTDIR

# Build for each os-architecture pair
for os in ${TARGET_OS} ; do
    archs="`f_build_archs ${os}`"
    exe_file_extension="`f_binary_file_ext ${os}`"
    for arch in ${archs} ; do
        make clean
        GOOS="${os}" GOARCH="${arch}" GOROOT_FINAL="$(go env GOROOT)" \
        make build \
            LDFLAGS=-buildid=${VERSION} \
            VERSION=${VERSION} \
            COMMIT=${COMMIT} \
            LEDGER_ENABLED=${LEDGER_ENABLED}
        mv ./build/${APP}${exe_file_extension} ${OUTDIR}/${APP}-${VERSION}-${os}-${arch}${exe_file_extension}
    done
    unset exe_file_extension
done

# Generate and display build report
f_generate_build_report ${OUTDIR}
cat ${OUTDIR}/build_report
