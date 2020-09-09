#/bin/bash

f_make_release_tarball() {
    local l_tarball l_dir

    l_dir=$1
    l_tarball=${l_dir}/${APP}-${VERSION}.tar.gz

    git archive --format tar.gz --prefix "${APP}-${VERSION}/" -o "${l_tarball}" HEAD

    l_tempdir="$(mktemp -d)"
    pushd "${l_tempdir}" >/dev/null
    tar xf "${l_tarball}"
    rm "${l_tarball}"
    find ${APP}-* | sort | tar --no-recursion --mode='u+rw,go+r-w,a+X' --owner=0 --group=0 -c -T - | gzip -9n > "${l_tarball}"
    popd >/dev/null
    rm -rf "${l_tempdir}"

    printf '%s' "${l_tarball}"
}

f_prepare_pristine_src_dir() {
    local l_tarball l_dir

    l_tarball=$1
    l_dir=$2

    pushd ${l_dir} >/dev/null
    tar --strip-components=1 -xf "${l_tarball}"
    go mod download
    popd >/dev/null
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
    local l_dir l_tempfile

    l_dir=$1
    l_tempfile="$(mktemp)"

    pushd "${l_dir}" >/dev/null
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
