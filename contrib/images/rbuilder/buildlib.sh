#/bin/bash

f_tarball_fix_file_order() {
    local l_tarball l_app l_dir

    l_tarball=$1
    l_app=$2
    l_dir="$(mktemp -d)"

    pushd "${l_dir}"
    tar xf "${l_tarball}"
    rm "${l_tarball}"
    find ${l_app}-* | sort | tar --no-recursion --mode='u+rw,go+r-w,a+X' --owner=0 --group=0 -c -T - | gzip -9n > "${l_tarball}"
    popd
    rm -rf "${l_dir}"
}

f_prepare_pristine_src_dir() {
    local l_tarball l_dir

    l_tarball=$1
    l_dir=$2

    pushd ${l_dir}
    tar --strip-components=1 -xf "${l_tarball}"
    go mod download
    popd
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
    local l_dir l_app l_version l_commit l_tempfile

    l_dir=$1
    l_app=$2
    l_version=$3
    l_commit=$4
    l_tempfile="$(mktemp)"

    pushd "${l_dir}"
    cat >>"${l_tempfile}" <<EOF
App: ${l_app}
Version: ${l_version}
Commit: ${l_commit}
EOF
    echo 'Files:' >> "${l_tempfile}"
    md5sum * | sed 's/^/ /' >> "${l_tempfile}"
    echo 'Checksums-Sha256:' >> "${l_tempfile}"
    sha256sum * | sed 's/^/ /' >> "${l_tempfile}"
    mv "${l_tempfile}" build_report
    popd
}
