#!/usr/bin/make -f

WAIT_TIME ?= 45s

all: test format

test:
	go test -mod=readonly -failfast -timeout=15m -tags='system_test' ./... --wait-time=$(WAIT_TIME) --verbose $(if $(findstring v2,$(COSMOS_BUILD_OPTIONS)),--binary=simdv2)

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofumpt -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gci write --skip-generated -s standard -s default -s "prefix(cosmossdk.io)" -s "prefix(github.com/cosmos/cosmos-sdk)" --custom-order

.PHONY: all test format
