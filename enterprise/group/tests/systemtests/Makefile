#!/usr/bin/make -f

WAIT_TIME ?= 45s

test:
	go test -failfast -timeout=15m -mod=readonly -tags='system_test' -v ./... --wait-time=$(WAIT_TIME) --binary simd --chain-id localnet --verbose
