PACKAGES=$(shell go list ./... | grep -v '/vendor/')
#BUILD_FLAGS = -ldflags "-X github.com/cosmos/cosmos-sdk/cmd/cosmos-sdk-cli/version.GitCommit=`git rev-parse --short HEAD`"

all: get_tools get_vendor_deps build test

get_tools:
	go get github.com/golang/dep/cmd/dep

build:
	go build -o bin/myawesomeprojectcli cmd/cli/main.go && go build -o bin/myawesomeprojectd cmd/node/main.go

get_vendor_deps:
	@rm -rf vendor/
	@dep ensure

test:
	@go test $(PACKAGES)

benchmark:
	@go test -bench=. $(PACKAGES)

.PHONY: all build test benchmark

