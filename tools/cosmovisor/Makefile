#!/usr/bin/make -f

all: cosmovisor test

cosmovisor:
	go build -mod=readonly ./cmd/cosmovisor
	@echo "cosmovisor binary has been successfully built in tools/cosmovisor/cosmovisor"

test:
	go test -mod=readonly -race ./...

.PHONY: all cosmovisor test
