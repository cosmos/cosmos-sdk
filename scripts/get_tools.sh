#!/usr/bin/env bash
set -e

# This file downloads all of the binary dependencies we have, and checks out a
# specific git hash.
#
# repos it installs:
#   github.com/golang/dep/cmd/dep
#   gopkg.in/alecthomas/gometalinter.v2
#   github.com/rakyll/statiik

## check if GOPATH is set
if [ -z ${GOPATH+x} ]; then
	echo "please set GOPATH (https://github.com/golang/go/wiki/SettingGOPATH)"
	exit 1
fi

mkdir -p "$GOPATH/src/github.com"
cd "$GOPATH/src/github.com" || exit 1

installFromGithub() {
	repo=$1
	commit=$2
	# optional
	subdir=$3
	echo "--> Installing $repo ($commit)..."
	if [ ! -d "$repo" ]; then
		mkdir -p "$repo"
		git clone "https://github.com/$repo.git" "$repo"
	fi
	if [ ! -z ${subdir+x} ] && [ ! -d "$repo/$subdir" ]; then
		echo "ERROR: no such directory $repo/$subdir"
		exit 1
	fi
	pushd "$repo" && \
		git fetch origin && \
		git checkout -q "$commit" && \
		if [ ! -z ${subdir+x} ]; then cd "$subdir" || exit 1; fi && \
		go install && \
		if [ ! -z ${subdir+x} ]; then cd - || exit 1; fi && \
		popd || exit 1
	echo "--> Done"
	echo ""
}

installFromGithub golang/dep 22125cfaa6ddc71e145b1535d4b7ee9744fefff2 cmd/dep
## gometalinter v2.0.11
installFromGithub alecthomas/gometalinter 17a7ffa42374937bfecabfb8d2efbd4db0c26741
installFromGithub rakyll/statik v0.1.5