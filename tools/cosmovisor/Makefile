#!/usr/bin/make -f

all: cosmovisor test

cosmovisor:
	go build -mod=ErrReadOnly ./cmd/cosmovisor

test:
	go test -mod=ErrReadOnly -race ./...

.PHONY: all cosmovisor test
