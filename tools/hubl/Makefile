#!/usr/bin/make -f

all: hubl test

hubl:
	go build -mod=ErrReadOnly ./cmd/hubl

test:
	go test -mod=ErrReadOnly -race ./...

.PHONY: all hubl test
