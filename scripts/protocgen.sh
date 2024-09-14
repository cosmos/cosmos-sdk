
#!/usr/bin/env bash

# How to run manually:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

echo "Formatting protobuf files"
find ./ -name "*.proto" -exec clang-format -i {} \;

set -e

home=$PWD

echo "Generating proto code"
proto_dirs=$(find ./ -name 'buf.yaml' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  echo "Generating proto code for $dir"

  cd "$dir"

  # Generate code using buf.gen.pulsar.yaml
  if [ -f "buf.gen.pulsar.yaml" ]; then
    buf generate --template buf.gen.pulsar.yaml

    # Move generated files to the correct place
    if [ -d "../cosmos" ] && [ "$dir" != "./proto" ]; then
      echo "Moving ../cosmos to $home/api"
      cp -r ../cosmos "$home/api"
      rm -rf ../cosmos
    fi
  fi

  # Generate code using buf.gen.gogo.yaml
  if [ -f "buf.gen.gogo.yaml" ]; then
    for file in $(find . -maxdepth 8 -name '*.proto'); do
      # Check if proto file has go_package set to cosmossdk.io/api and should not use gogo proto
      if grep -q "option go_package" "$file" && ! grep -q 'option go_package.*cosmossdk.io/api' "$file"; then
        echo "Generating gogo proto code for $file"
        buf generate --template buf.gen.gogo.yaml "$file"
      fi
    done

    # Move generated files to the right places
    if [ -d "../cosmossdk.io" ]; then
      echo "Moving ../cosmossdk.io files to $home"
      cp -r ../cosmossdk.io/* "$home"
      rm -rf ../cosmossdk.io
    fi

    if [ -d "../github.com" ] && [ "$dir" != "./proto" ]; then
      echo "Moving ../github.com/cosmos/cosmos-sdk files to $home"
      cp -r ../github.com/cosmos/cosmos-sdk/* "$home"
      rm -rf ../github.com
    fi
  fi

  # Return to the home directory
  cd "$home"
done

# Move final generated files to the current directory
if [ -d "github.com/cosmos/cosmos-sdk" ]; then
  echo "Moving final generated files from github.com/cosmos/cosmos-sdk to $home"
  cp -r github.com/cosmos/cosmos-sdk/* ./
  rm -rf github.com
fi

# Cleaning up certain files to avoid generation issues with Pulsar
# rm -r api/cosmos/bank/v2
# rm -r api/cosmos/bank/module/v2

# Tidy up Go modules
go mod tidy