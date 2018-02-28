PACKAGES=$(shell go list ./... | grep -v '/vendor/')
BUILD_FLAGS = -ldflags "-X github.com/cosmos/cosmos-sdk/examples/basecoin/version.GitCommit=`git rev-parse --short HEAD`"

all: build test

build:
	go build $(BUILD_FLAGS) -o build/basecoin ./cmd/...

test:
	@go test $(PACKAGES)

benchmark:
	@go test -bench=. $(PACKAGES)

.PHONY: all build test benchmark
