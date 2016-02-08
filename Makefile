.PHONY: all test get_deps

all: test install

install: get_deps
	go install github.com/tendermint/blackstar/cmd/...

test:
	go test github.com/tendermint/blackstar/...

get_deps:
	go get -d github.com/tendermint/blackstar/...
