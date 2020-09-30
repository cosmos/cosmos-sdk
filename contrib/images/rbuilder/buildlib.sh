#/bin/bash

f_make_release_tarball() {
    SOURCEDIST=${BASEDIR}/${APP}-${VERSION}.tar.gz

    git archive --format tar.gz --prefix "${APP}-${VERSION}/" -o "${SOURCEDIST}" HEAD

    l_tempdir="$(mktemp -d)"
    pushd "${l_tempdir}" >/dev/null
    tar xf "${SOURCEDIST}"
    rm "${SOURCEDIST}"
    find ${APP}-* | sort | tar --no-recursion --mode='u+rw,go+r-w,a+X' --owner=0 --group=0 -c -T - | gzip -9n > "${SOURCEDIST}"
    popd >/dev/null
    rm -rf "${l_tempdir}"
}

f_setup_pristine_src_dir() {
    cd ${pristinesrcdir}
    tar --strip-components=1 -xf "${SOURCEDIST}"
    go mod download
}

f_build_archs() {
    local l_os

    l_os=$1

    case "${l_os}" in
    darwin | windows)
        echo 'amd64'
        ;;
    linux)
        echo 'amd64 arm64'
        ;;
    *)
        echo "unknown OS -- ${l_os}" >&2
        return 1
    esac
}

f_binary_file_ext() {
   [ $1 = windows ] && printf '%s' '.exe' || printf ''
}

f_generate_build_report() {
    local l_tempfile

    l_tempfile="$(mktemp)"

    pushd "${OUTDIR}" >/dev/null
    cat >>"${l_tempfile}" <<EOF
App: ${APP}
Version: ${VERSION}
Commit: ${COMMIT}
EOF
    echo 'Files:' >> "${l_tempfile}"
    md5sum * | sed 's/^/ /' >> "${l_tempfile}"
    echo 'Checksums-Sha256:' >> "${l_tempfile}"
    sha256sum * | sed 's/^/ /' >> "${l_tempfile}"
    mv "${l_tempfile}" build_report
    popd >/dev/null
}

[ "x${DEBUG}" = "x" ] || set -x

BASEDIR="$(mktemp -d)"
OUTDIR=$HOME/artifacts
rm -rfv ${OUTDIR}/
mkdir -p ${OUTDIR}/
pristinesrcdir=${BASEDIR}/buildsources
mkdir -p ${pristinesrcdir}

# Make release tarball
f_make_release_tarball

# Extract release tarball and cache dependencies
f_setup_pristine_src_dir

# Move the release tarball to the out directory
mv ${SOURCEDIST} ${OUTDIR}/
