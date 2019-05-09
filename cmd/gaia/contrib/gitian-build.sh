#!/bin/bash

# symbol prefixes:
# g_ -> global
# l_ - local variable
# f_ -> function

set -euo pipefail

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
THIS="${THIS_DIR}/$(basename ${BASH_SOURCE[0]})"
GITIAN_CACHE_DIRNAME='.gitian-builder-cache'
GO_TARBALL='golang-debian-1.12.4-1.tar.gz'

# Defaults

DEFAULT_SIGN_COMMAND='gpg --detach-sign'
DEFAULT_GAIA_SIGS=${GAIA_SIGS:-'gaia.sigs'}
DEFAULT_GITIAN_REPO='https://github.com/devrandom/gitian-builder'
DEFAULT_GBUILD_FLAGS=''

# Overrides

SIGN_COMMAND=${SIGN_COMMAND:-${DEFAULT_SIGN_COMMAND}}
GITIAN_REPO=${GITIAN_REPO:-${DEFAULT_GITIAN_REPO}}
GBUILD_FLAGS=${GBUILD_FLAGS:-${DEFAULT_GBUILD_FLAGS}}

# Globals

g_workdir=''
g_gitian_cache=''
g_sign_identity=''
g_gitian_skip_download=''

f_main() {
  local l_dirname \
    l_sdk \
    l_commit \
    l_platform \
    l_result \
    l_descriptor \
    l_release \
    l_sigs_dir

  l_platform=$1
  l_sdk=$2
  l_sigs_dir=$3

  pushd ${l_sdk}
  l_commit="$(git rev-parse HEAD)"
  l_release="$(git describe --tags | sed 's/^v//')-${l_platform}"
  popd

  l_descriptor=${THIS_DIR}/gitian-descriptors/gitian-${l_platform}.yml
  [ -f ${l_descriptor} ]

  if [ "${g_gitian_skip_download}" != "y" ]; then
    echo "Download gitian-builder to ${g_workdir}" >&2
    git clone ${GITIAN_REPO} ${g_workdir}
  fi

  echo "Fetching Go sources" >&2
  f_ensure_go_source_tarball

  echo "Prepare gitian-target docker image" >&2
  f_prep_docker_image

  echo "Start the build" >&2
  f_build "${l_descriptor}" "${l_commit}"
  echo "You may find the result in $(echo ${g_workdir}/result/*.yml))" >&2

  if [ -n "${g_sign_identity}" ]; then
    f_sign "${l_descriptor}" "${l_release}" "${l_sigs_dir}"
    echo "Build signed as ${g_sign_identity}, signatures can be found in ${l_sigs_dir}"
    f_verify "${l_descriptor}" "${l_release}" "${l_sigs_dir}"
    echo "Signatures in ${l_sigs_dir} have been verified"
  else
    echo "You can now sign the build with the following command:" >&2
    echo "cd ${g_workdir} ; bin/gsign -p 'gpg --detach-sign --armor' -s GPG_IDENTITY --release=${l_release} ${l_descriptor}" >&2
  fi

  return 0
}

f_prep_docker_image() {
  pushd ${g_workdir}
  bin/make-base-vm --docker --suite bionic --arch amd64
  popd
}

f_ensure_go_source_tarball() {
  local l_cached_tar

  l_cached_tar="${g_gitian_cache}/${GO_TARBALL}"
  mkdir -p ${g_workdir}/inputs
  if [ -f "${l_cached_tar}" ]; then
    echo "Fetching cached tarball from ${l_cached_tar}" >&2
    cp "${l_cached_tar}" ${g_workdir}/inputs/${GO_TARBALL}
  else
    f_download_go
    echo "Caching ${GO_TARBALL}" >&2
    cp ${g_workdir}/inputs/${GO_TARBALL} "${l_cached_tar}"
  fi
}

f_download_go() {
  local l_remote

  l_remote=https://salsa.debian.org/go-team/compiler/golang/-/archive/debian/1.12.4-1
  curl -L "${l_remote}/${GO_TARBALL}" > ${g_workdir}/inputs/${GO_TARBALL}
}

f_build() {
  local l_sdk l_descriptor

  l_descriptor=$1
  l_commit=$2

  [ -f ${l_descriptor} ]

  cd ${g_workdir}
  export USE_DOCKER=1
  bin/gbuild --commit cosmos-sdk="$l_commit" ${GBUILD_FLAGS} "$l_descriptor"
  libexec/stop-target || echo "warning: couldn't stop target" >&2
}

f_sign() {
  local l_descriptor l_release_name l_sigs_dir

  l_descriptor=$1
  l_release_name=$2
  l_sigs_dir=$3

  pushd ${g_workdir}
  bin/gsign -p "${SIGN_COMMAND}" -s "${g_sign_identity}" --destination="${l_sigs_dir}" --release=${l_release_name} ${l_descriptor}
  popd
}

f_verify() {
  local l_descriptor l_release_name l_sigs_dir

  l_descriptor=$1
  l_release_name=$2
  l_sigs_dir=$3

  pushd ${g_workdir}
  bin/gverify --destination="${l_sigs_dir}" --release="${l_release_name}" ${l_descriptor}
  popd
}

f_validate_platform() {
  case "${1}" in
  linux|darwin|windows)
    ;;
  *)
    echo "invalid platform -- ${1}"
    exit 1
  esac
}

f_abspath() {
  echo "$(cd "$(dirname "$1")"; pwd -P)/$(basename "$1")"
}

f_help() {
  cat >&2 <<EOF
Usage: $(basename $0) [-h] GOOS GIT_REPO
Launch a gitian build from the local clone of cosmos-sdk available at GIT_REPO.

  Options:
   -h               display this help and exit
   -d DIRNAME       set working directory name and skip gitian-builder download
   -s IDENTITY      sign build as IDENTITY

If a GPG identity is supplied via the -s flag, the build will be signed and verified.
The signature will be saved in '${DEFAULT_GAIA_SIGS}/'. An alternative output directory
for signatures can be supplied via the environment variable \$GAIA_SIGS.

The default signing command used to sign the build is '$DEFAULT_SIGN_COMMAND'.
An alternative signing command can be supplied via the environment
variable \$SIGN_COMMAND.
EOF
}

while getopts ":d:s:h" opt; do
  case "${opt}" in
    h)  f_help ; exit 0 ;;
    d)  g_dirname="${OPTARG}" ;;
    s)  g_sign_identity="${OPTARG}" ;;
  esac
done

shift "$((OPTIND-1))"

g_platform="${1}"
f_validate_platform "${g_platform}"

g_dirname="${g_dirname:-gitian-build-${g_platform}}"
g_workdir="$(pwd)/${g_dirname}"
if [ -d "${g_workdir}" ]; then
  echo "Directory ${g_workdir} exists and will be preserved" >&2
  g_gitian_skip_download=y
fi

g_sdk="$(f_abspath ${2})"
[ -d "${g_sdk}" ]

g_sigs_dir=${GAIA_SIGS:-"$(pwd)/${DEFAULT_GAIA_SIGS}"}

# create local cache directory for all gitian builds
g_gitian_cache="$(pwd)/${GITIAN_CACHE_DIRNAME}"
echo "Ensure cache directory ${g_gitian_cache} exists" >&2
mkdir -p "${g_gitian_cache}"

f_main "${g_platform}" "${g_sdk}" "${g_sigs_dir}"
