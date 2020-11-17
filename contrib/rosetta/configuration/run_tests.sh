#!/bin/sh

wait_for_rosetta() {
  timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' rosetta 8080
}

wait_for_rosetta

rosetta-cli check:data --configuration-file rosetta.json