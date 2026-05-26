#!/bin/sh

set -ex

# Usage: protogen.sh <module>
#   module: enterprise module name (group or poa)
# Must be run from the module directory (where proto/ and x/ exist).
#
# Uses the same buf/proto-builder tooling as scripts/protocgen.sh.
# Enterprise modules have a different structure (buf.gen in proto/, different output paths)
# so we cannot use the main protocgen.sh directly.

MODULE="$1"

if [ -z "$MODULE" ]; then
    echo "Usage: protogen.sh <module>"
    echo "  module: group or poa"
    exit 1
fi

# Run buf generate - same approach as main SDK (scripts/protocgen.sh, protocgen-pulsar.sh)
cd proto
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..

# Module-specific: copy generated files to the right places
case "$MODULE" in
    group)
        cp ./github.com/cosmos/cosmos-sdk/enterprise/group/*.pb.go ./x/group/ 2>/dev/null || true
        cp ./github.com/cosmos/cosmos-sdk/enterprise/group/*.pb.gw.go ./x/group/ 2>/dev/null || true
        rm -rf ./github.com ./cosmos
        ;;
    poa)
        cp -r ./github.com/cosmos/cosmos-sdk/enterprise/poa/types/* ./x/poa/types
        rm -rf ./github.com
        ;;
    *)
        echo "Unknown module: $MODULE (expected group or poa)"
        exit 1
        ;;
esac

# Add license headers to all generated files
# When run in docker, /repo is the cosmos-sdk root; when run locally, derive from cwd
if [ -d "/repo/enterprise/scripts" ]; then
    REPO_ROOT="/repo"
else
    REPO_ROOT="$(cd ../.. && pwd)"
fi
echo "Adding license headers to generated files..."
sh "$REPO_ROOT/enterprise/scripts/add-license.sh" "$(pwd)" "enterprise/$MODULE"
