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
DISTNAME=${APP}-${VERSION}
SOURCEDIST=${BASEDIR}/${DISTNAME}.tar.gz
OUTDIR=$HOME/artifacts
rm -rfv ${OUTDIR}/
mkdir -p ${OUTDIR}/
pristinesrcdir=${BASEDIR}/buildsources
mkdir -p ${pristinesrcdir}

# Make release tarball
git archive --format tar.gz --prefix ${DISTNAME}/ -o ${SOURCEDIST} HEAD

# Source builder's functions library
. /usr/local/share/cosmos-sdk/buildlib.sh

# Correct tar file order
f_tarball_fix_file_order ${SOURCEDIST} ${APP}

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
        make simd \
            LDFLAGS=-buildid=${VERSION} \
            VERSION=${VERSION} \
            COMMIT=${COMMIT} \
            LEDGER_ENABLED=${LEDGER_ENABLED}
        mv ./build/simd${exe_file_extension} ${OUTDIR}/${DISTNAME}-${os}-${arch}${exe_file_extension}
    done
    unset exe_file_extension
done

f_generate_build_report ${OUTDIR} ${APP} ${VERSION} ${COMMIT}
cat ${OUTDIR}/build_report
