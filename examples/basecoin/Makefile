PACKAGES=$(shell go list ./... | grep -v '/vendor/')
BUILD_FLAGS = -ldflags "-X github.com/cosmos/cosmos-sdk/examples/basecoin/version.GitCommit=`git rev-parse --short HEAD`"

all: get_vendor_deps build test 

build:
	go build $(BUILD_FLAGS) -o build/basecoin ./cmd/...

get_vendor_deps:
	@rm -rf vendor/
	@glide install

test:
	@go test $(PACKAGES)

benchmark:
	@go test -bench=. $(PACKAGES)

.PHONY: build get_vendor_deps test benchmark 
