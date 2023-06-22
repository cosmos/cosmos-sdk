#!/usr/bin/make -f

all: stateviewer test

stateviewer:
	go build -mod=readonly ./cmd/stateviewer

test:
	go test -mod=readonly -race ./...

.PHONY: all stateviewer test
