#!/bin/sh

set -e

build() {
    echo finding protobuf files in "$1"
    proto_files=$(find "$1" -name "*.proto")
    for file in $proto_files; do
      echo "building proto file $file"
      protoc -I=. -I=./third_party/proto --plugin /usr/bin/protoc-gen-go-cosmos-orm --go-cosmos-orm_out=.  --go_out=. "$file"
    done
}

for dir in "$@"
do
  build "$dir"
done

cp -r github.com/cosmos/cosmos-sdk/orm/* ./
rm -rf github.com