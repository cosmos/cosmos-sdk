#!/usr/bin/make -f

all: confix condiff test

confix:
	go build -mod=readonly ./cmd/confix
	@echo "confix binary has been successfully built in tools/confix/confix"

condiff:
	go build -mod=readonly ./cmd/condiff
	@echo "condiff binary has been successfully built in tools/confix/condiff"

test:
	go test -mod=readonly -race ./...

.PHONY: all confix condiff test
