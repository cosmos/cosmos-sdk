#!/usr/bin/make -f

all: build

rosetta:
	go build -mod=ErrReadOnly ./cmd/rosetta

build:
	go build ./cmd/rosetta.go

test:
	go test -mod=ErrReadOnly -race ./...

.PHONY: all build rosetta test