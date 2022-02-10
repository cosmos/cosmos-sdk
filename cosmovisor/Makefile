#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --always --match "cosmovisor/v*") | sed 's/^cosmovisor[/]//')

all: cosmovisor test

cosmovisor:
	go build -ldflags="-X 'github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd.Version=$(VERSION)'" -mod=readonly ./cmd/cosmovisor

test:
	go test -mod=readonly -race ./...

.PHONY: all cosmovisor test
