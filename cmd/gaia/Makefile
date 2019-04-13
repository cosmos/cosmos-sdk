#!/usr/bin/make -f

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
GOBIN ?= $(GOPATH)/bin
GOSUM := $(shell which gosum)

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags)"

ifneq ($(GOSUM),)
ldflags += -X github.com/cosmos/cosmos-sdk/version.GoSumHash=$(shell $(GOSUM) ../../go.sum)
endif

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

# The below include contains the tools target.
include ../../contrib/devtools/Makefile

all: install lint check

build: ../../go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/gaiad.exe ../../cmd/gaia/cmd/gaiad
	go build -mod=readonly $(BUILD_FLAGS) -o build/gaiacli.exe ../../cmd/gaia/cmd/gaiacli
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/gaiad ../../cmd/gaia/cmd/gaiad
	go build -mod=readonly $(BUILD_FLAGS) -o build/gaiacli ../../cmd/gaia/cmd/gaiacli
	go build -mod=readonly $(BUILD_FLAGS) -o build/gaiareplay ../../cmd/gaia/cmd/gaiareplay
	go build -mod=readonly $(BUILD_FLAGS) -o build/gaiakeyutil ../../cmd/gaia/cmd/gaiakeyutil
endif

build-linux: ../../go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

install: ../../go.sum check-ledger
	go install -mod=readonly $(BUILD_FLAGS) ../../cmd/gaia/cmd/gaiad
	go install -mod=readonly $(BUILD_FLAGS) ../../cmd/gaia/cmd/gaiacli
	go install -mod=readonly $(BUILD_FLAGS) ../../cmd/gaia/cmd/gaiareplay
	go install -mod=readonly $(BUILD_FLAGS) ../../cmd/gaia/cmd/gaiakeyutil

install-debug: ../../go.sum
	go install -mod=readonly $(BUILD_FLAGS) ../../cmd/gaia/cmd/gaiadebug


########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/gaiad -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

########################################
### Testing


check: check-unit
check-unit:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

check-race:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' -race ./...

check-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

check-build: build
	@go test -mod=readonly -p 4 `go list ./cli_test/...` -tags=cli_test

check-all: check-unit check-race check-cover check-build

lint: ci-lint
ci-lint:
	golangci-lint run
	go vet -composites=false -tests=false ./...
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/cosmos/cosmos-sdk

benchmark:
	@go test -mod=readonly -bench=. ./...


# include simulations
include sims.mk

.PHONY: all build-linux install install-debug \
	go-mod-cache draw-deps clean \
	check check-all check-build check-cover check-ledger check-unit check-race

