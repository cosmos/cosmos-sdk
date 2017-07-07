#! /bin/bash 

# Get the version from the environment, or try to figure it out.
if [ -z $VERSION ]; then
    VERSION=$(awk -F\" '/Version =/ { print $2; exit }' < version/version.go)
fi
aws s3 cp --recursive build/dist s3://tendermint/binaries/basecoin/v${VERSION} --acl public-read
