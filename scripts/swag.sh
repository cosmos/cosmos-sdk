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

${swag_cmd} init -g client/rest/root.go --output client/rest/docs

if (($(git status --porcelain 2>/dev/null | grep 'client/rest/docs/swagger.yaml' | wc -l) > 0)); then
  echo -e "${RED}Swagger docs are out of sync!!!${NC}"
  exit 1
else
  echo -e "${GREEN}Swagger docs are in sync${NC}"
fi
