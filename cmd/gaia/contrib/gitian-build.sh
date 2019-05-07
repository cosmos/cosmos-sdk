#!/bin/bash

set -euo pipefail

SIGN_COMMAND=${SIGN_COMMAND:-'gpg --detach-sign --armor'}

g_workdir=''
g_sign_identity=''

f_main() {
  local l_dirname \
    l_sdk \
    l_commit \
    l_platform \
    l_result \
    l_descriptor \
    l_release

  l_platform=$1
  l_sdk=$2
  l_release=$3
  pushd ${l_sdk}
  l_commit=$(git rev-parse HEAD)
  popd

  l_descriptor=$l_sdk/cmd/gaia/contrib/gitian-descriptors/gitian-${l_platform}.yml
  [ -f ${l_descriptor} ]

  echo "Download gitian" >&2
  git clone https://github.com/devrandom/gitian-builder ${g_workdir}

  echo "Prepare gitian-target docker image" >&2
  f_prep_docker_image

  echo "Download Go" >&2
  f_download_go

  echo "Start the build" >&2
  f_build "${l_sdk}" "${l_commit}" "${l_platform}"
  echo "You may find the result in $(echo ${g_workdir}/result/*.yml))" >&2

  if [ -n "${g_sign_identity}" ]; then
    f_sign ${l_descriptor} ${l_release}
    echo "Build signed as ${g_sign_identity}: ${g_workdir}/sigs/"
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

f_download_go() {
  local l_gopkg

  l_gopkg=go1.12.4.linux-amd64.tar.gz

  mkdir -p ${g_workdir}/inputs
  curl -L https://dl.google.com/go/${l_gopkg} > ${g_workdir}/inputs/${l_gopkg}
}

f_build() {
  local l_sdk l_platform l_descriptor

  l_sdk=$1
  l_commit=$2
  l_platform=$3
  l_descriptor=$l_sdk/cmd/gaia/contrib/gitian-descriptors/gitian-${l_platform}.yml

  [ -f ${l_descriptor} ]

  cd ${g_workdir}
  export USE_DOCKER=1
  bin/gbuild $l_descriptor --commit cosmos-sdk=$l_commit
  libexec/stop-target || echo "warning: couldn't stop target" >&2
}

f_sign() {
  local l_descriptor l_release_name

  l_descriptor=$1
  l_release_name=$2

  cd ${g_workdir}
  bin/gsign -p "${SIGN_COMMAND}" -s "${g_sign_identity}" --release=${l_release_name} ${l_descriptor}
}

f_validate_platform() {
  case "${1}" in
  linux|osx|windows)
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
Usage: $(basename $0) [-h] GOOS GIT_REPO MAJ_MIN_RELEASE
Launch a gitian build from the local clone of cosmos-sdk available at GIT_REPO.

  Options:
   -h               display this help and exit
   -d DIRNAME       set working directory name
   -s IDENTITY      sign build as IDENTITY

The default signing command used to sign the build is '$SIGN_COMMAND'.
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
mkdir "${g_workdir}"

g_sdk="$(f_abspath ${2})"
[ -d "${g_sdk}" ]

f_main "${g_platform}" "${g_sdk}" "${3}"
