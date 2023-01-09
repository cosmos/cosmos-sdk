#!/usr/bin/make -f

all: build

rosetta:
	go build -mod=readonly ./cmd/rosetta

build:
	go build ./cmd/rosetta.go

test:
	go test -mod=readonly -race ./...

.PHONY: all build rosetta test