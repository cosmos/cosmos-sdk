#!/usr/bin/make -f

all: confix condiff test

confix:
	go build -mod=ErrReadOnly ./cmd/confix

condiff:
	go build -mod=ErrReadOnly ./cmd/condiff

test:
	go test -mod=ErrReadOnly -race ./...

.PHONY: all confix condiff test
