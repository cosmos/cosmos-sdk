#!/usr/bin/env bash

set -eo pipefail

pwd=$(pwd)

mkdir -p ./tmp-swagger-gen
for i in `find . -name buf.gen.swagger.yaml -exec realpath {} \; 2> /dev/null`
do
  dir=$(dirname $i)
  cd $dir
  proto_dirs=$(find $dir -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
  for dir in $proto_dirs; do
    echo "Generating swagger files for $dir"
    # generate swagger files (filter query files)
    query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
    if [[ ! -z "$query_file" ]]; then
      buf generate --template buf.gen.swagger.yaml $query_file
    fi
  done
done

cd $pwd

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
