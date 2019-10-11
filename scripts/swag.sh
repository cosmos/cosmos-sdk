#!/bin/bash

swag_cmd=$1
if [ -z "$swag_cmd" ]
then
  echo "must provide a valid path to the swag binary"
  exit 1
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Generate Swagger documentation from source and verify annotations have not
# changed by checking if the swagger.yaml config is dirty or not. Since the
# 'swag init' command does not currently support ignoring timestamps, the doc.go
# file will always be modified so we must prevent that from always becoming dirty
# by replacing it with the original after the command is executed.
cp client/rest/docs/docs.go client/rest/docs/docs_bak.go
${swag_cmd} init -g client/rest/root.go --output client/rest/docs
mv client/rest/docs/docs_bak.go client/rest/docs/docs.go

if (($(git status --porcelain 2>/dev/null | grep 'client/rest/docs/swagger.yaml' | wc -l) > 0)); then
  echo -e "${RED}Swagger docs are out of sync!!!${NC}"
  exit 1
else
  echo -e "${GREEN}Swagger docs are in sync${NC}"
fi
