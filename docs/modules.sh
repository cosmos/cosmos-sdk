#!/usr/bin/env bash

rm -rf modules
mkdir modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    mkdir -p "modules/$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/spec/ "$_"
  fi
done