#!/usr/bin/make -f

WAIT_TIME ?= 20s

# Build the transitional POA simapp binary (post-upgrade binary with migration handler).
build-poa:
	cd ../../examples/migrate-from-pos && go build -o ../../tests/systemtests/binaries/simd ./simd

# Build the standard SDK simapp binary (pre-upgrade POS binary).
build-pos:
	cd ../../../../simapp && go build -o ../../../../enterprise/poa/tests/systemtests/binaries/pos/simd ./simd

# Build both binaries.
build: build-poa build-pos

test: build
	go test -failfast -timeout=30m -mod=readonly -tags='system_test' -v --wait-time=$(WAIT_TIME) --binary simd --chain-id localnet

# Run only the upgrade test.
test-upgrade: build
	go test -failfast -timeout=30m -mod=readonly -tags='system_test' -v -run TestPOStoPoaUpgrade --wait-time=$(WAIT_TIME) --binary simd --chain-id localnet
