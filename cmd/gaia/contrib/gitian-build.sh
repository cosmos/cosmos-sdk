#!/bin/bash

set -euo pipefail

g_default_dirname="gitian-build-$(date +%Y-%m-%d-%H%M%S)"
g_dirname=''

f_main() {
  local l_dirname \
    l_workdir \
    l_sdk \
    l_commit \
    l_platform \
    l_result \
    l_descriptor \
    l_release_name

  l_dirname=$1
  l_platform=$2
  l_sdk=$3
  pushd ${l_sdk}
  l_commit=$(git rev-parse HEAD)
  popd

  l_descriptor=$l_sdk/cmd/gaia/contrib/gitian-descriptors/gitian-${l_platform}.yml
  [ -f ${l_descriptor} ]

  l_workdir="$(pwd)/${l_dirname}"
  mkdir ${l_workdir}/
  echo "Download gitian" >&2
  git clone https://github.com/devrandom/gitian-builder ${l_workdir}

  echo "Prepare gitian-target docker image" >&2
  f_prep_docker_image "${l_workdir}"

  echo "Download Go" >&2
  f_download_go "${l_workdir}"

  echo "Start the build" >&2
  f_build "${l_workdir}" "${l_sdk}" "${l_commit}" "${l_platform}"
  echo "You may find the result in $(echo ${l_workdir}/result/*.yml))" >&2

  l_release_name="$(sed -n 's/^name: \"\(.\+\)\"$/\1/p' ${l_descriptor})"
  [ -n ${l_release_name} ]
  echo "You can now sign the build with the following command:" >&2
  echo "cd ${l_workdir} ; bin/gsign -p 'gpg --detach-sign --armor' -s GPG_IDENTITY --release=${l_release_name} ${l_descriptor}" >&2
  return 0
}

f_prep_docker_image() {
  local l_dir=$1

  pushd ${l_dir}
  bin/make-base-vm --docker --suite bionic --arch amd64
  popd
}

f_download_go() {
  local l_dir l_gopkg

  l_dir=$1

  mkdir -p ${l_dir}/inputs
  l_gopkg=go1.12.4.linux-amd64.tar.gz
  curl -L https://dl.google.com/go/${l_gopkg} > ${l_dir}/inputs/${l_gopkg}
}

f_build() {
  local l_gitian l_sdk l_platform l_descriptor

  l_gitian=$1
  l_sdk=$2
  l_commit=$3
  l_platform=$4
  l_descriptor=$l_sdk/cmd/gaia/contrib/gitian-descriptors/gitian-${l_platform}.yml

  [ -f ${l_descriptor} ]

  cd ${l_gitian}
  export USE_DOCKER=1
  bin/gbuild $l_descriptor --commit cosmos-sdk=$l_commit
  libexec/stop-target || echo "warning: couldn't stop target" >&2
}

f_help() {
  cat >&2 <<EOF
Usage: $(basename $0) [-h] GOOS GIT_REPO
Launch a gitian build from the local clone of cosmos-sdk available at GIT_REPO.

  Options:
   -h               Display this help and exit
   -d DIRNAME       Set working directory name
EOF
}

while getopts ":d:h" opt; do
  case "${opt}" in
    h)  f_help ; exit 0 ;;
    d)  g_dirname="${OPTARG}"
  esac
done

shift "$((OPTIND-1))"

f_main "${g_dirname:-${g_default_dirname}}" $1 $2
