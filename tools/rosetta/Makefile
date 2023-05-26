#!/usr/bin/make -f

all: build

rosetta:
	go build -mod=readonly -buildmode=plugin ./cmd/rosetta

test:
	go test -mod=readonly -race ./...

build:
	cd plugins/default && make plugin
	cd plugins/osmosis && make plugin
	go build -mod=readonly -buildmode=plugin ./cmd/rosetta

.PHONY: all build rosetta test