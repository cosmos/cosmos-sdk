#!/usr/bin/make -f

all: hubl test

hubl:
	go build -mod=readonly ./cmd/hubl
	@echo "hubl binary has been successfully built in tools/hubl/hubl"

test:
	go test -mod=readonly -race ./...

.PHONY: all hubl test
