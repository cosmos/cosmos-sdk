#!/bin/bash

if [ $# -lt 3 ]; then
    echo "Usage: check-compat.sh <branch> <simapp_version> [<go_mod_name1> <go_mod_name2> ...]"
    exit 1
fi

dir="tmp"
branch=$1
simapp_version=$2
shift 3
go_mod_names=("$@")

# clone cosmos-sdk
export FILTER_BRANCH_SQUELCH_WARNING=1
git clone -b $branch --depth 1 https://github.com/cosmos/cosmos-sdk $dir

# save last commit branch commit
COMMIT=$(git rev-parse HEAD)
# save the last main commit
latest_commit=$(git ls-remote https://github.com/cosmos/cosmos-sdk.git refs/heads/main | cut -f1 || "main")

# if simapp_version is v2 then use simapp/v2
if [ $simapp_version == "v2" ]; then
  cd $dir/simapp/v2
else
  cd $dir/simapp
fi

# bump all cosmos-sdk packages to latest branch commit
VERSIONS=$(go mod edit -json | jq -r '.Replace[].Old.Path')

# Initialize variables for different types of replaces
BRANCH_REPLACES=""
MAIN_REPLACES=""
REQUIRES=""

for version in $VERSIONS; do
   if [[ " ${go_mod_names[@]} " =~ " ${version} " ]]; then
    MAIN_REPLACES+=" -replace $version=$version@$latest_commit"
    continue
  elif [[ $version == "github.com/cosmos/cosmos-sdk"* || $version == "cosmossdk.io/"* ]]; then
    BRANCH_REPLACES+=" -replace $version=$version@$COMMIT"
  fi
done

for mod in ${go_mod_names[@]}; do
  REQUIRES+=" -require $mod@$latest_commit"
done

# Apply the replaces
go mod edit $BRANCH_REPLACES $MAIN_REPLACES $REQUIRES

go mod tidy

# Test SimApp
go test -mod=readonly -v ./...