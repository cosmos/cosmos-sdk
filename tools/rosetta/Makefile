#!/usr/bin/make -f

all: build

rosetta:
	go build -mod=readonly - ./cmd/rosetta

build:
	go build ./cmd/rosetta.go

test:
	go test -mod=readonly -race ./...

kek:
	cd plugins/osmosis && go mod tidy && make plugin
	go run cmd/rosetta/main.go --blockchain "osmosis" --denom-to-suggest "osmo" --network "osmo-test-4" --tendermint "tcp://127.0.0.1:26657" --grpc "lola.wk.zondax.ch:9090"

.PHONY: all build rosetta test