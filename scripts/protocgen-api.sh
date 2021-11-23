#!/usr/bin/env bash

# this script generates the new API go module using pulsar

set -eo pipefail

cd api
buf generate .
