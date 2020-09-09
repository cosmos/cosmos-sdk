#!/bin/bash

set -ue

# Expect the following envvars to be set:
# - APP
# - VERSION
# - COMMIT
# - TARGET_OS
# - LEDGER_ENABLED
# - DEBUG

[ "x${DEBUG}" = "x" ] || set -x

BASEDIR="$(mktemp -d)"
OUTDIR=$HOME/artifacts
rm -rfv ${OUTDIR}/
mkdir -p ${OUTDIR}/
pristinesrcdir=${BASEDIR}/buildsources
mkdir -p ${pristinesrcdir}


# Make release tarball
SOURCEDIST="`f_make_release_tarball ${BASEDIR}`"

# Extract release tarball and cache dependencies
f_prepare_pristine_src_dir "${SOURCEDIST}" "${pristinesrcdir}"

# Move the release tarball to the out directory
mv ${SOURCEDIST} ${OUTDIR}/

# Build for each os-architecture pair
cd ${pristinesrcdir}
for os in ${TARGET_OS} ; do
    archs="`f_build_archs ${os}`"
    exe_file_extension="`f_binary_file_ext ${os}`"
    for arch in ${archs} ; do
        make clean
        GOOS="${os}" GOARCH="${arch}" GOROOT_FINAL="$(go env GOROOT)" \
        make ${APP} \
            LDFLAGS=-buildid=${VERSION} \
            VERSION=${VERSION} \
            COMMIT=${COMMIT} \
            LEDGER_ENABLED=${LEDGER_ENABLED}
        mv ./build/${APP}${exe_file_extension} ${OUTDIR}/${APP}-${VERSION}-${os}-${arch}${exe_file_extension}
    done
    unset exe_file_extension
done

f_generate_build_report ${OUTDIR}
cat ${OUTDIR}/build_report
