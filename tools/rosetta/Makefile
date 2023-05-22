#!/usr/bin/make -f

all: build

plugin:
	go build -buildmode=plugin -o plugins/osmosis.so plugins/osmosis.go

rosetta:
	go build -mod=readonly - ./cmd/rosetta

build:
	go build ./cmd/rosetta.go

test:
	go test -mod=readonly -race ./...

.PHONY: all build rosetta test