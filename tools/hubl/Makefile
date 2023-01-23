#!/usr/bin/make -f

all: hubl test

hubl:
	go build -mod=readonly ./cmd/hubl

test:
	go test -mod=readonly -race ./...

.PHONY: all hubl test
