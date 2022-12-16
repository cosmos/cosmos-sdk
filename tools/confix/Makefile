#!/usr/bin/make -f

all: confix test

confix:
	go build -mod=readonly ./cmd/confix

test:
	go test -mod=readonly -race ./...

.PHONY: all confix test
