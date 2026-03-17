#!/usr/bin/make -f

WAIT_TIME ?= 45s

all: test format

# test-short: ~5–7 min, for CI on PRs. Skips load, upgrade, stability, pruning/retention tests.
test-short:
	go test -mod=readonly -failfast -short -timeout=7m -v -tags='system_test' ./... --wait-time=$(WAIT_TIME) --verbose

# test: full suite (~30 min), for local runs and merges to main
test:
	go test -mod=readonly -failfast -timeout=30m -v -tags='system_test' ./... --wait-time=$(WAIT_TIME) --verbose

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofumpt -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gci write --skip-generated -s standard -s default -s "prefix(cosmossdk.io)" -s "prefix(github.com/cosmos/cosmos-sdk)" --custom-order

.PHONY: all test test-short format
