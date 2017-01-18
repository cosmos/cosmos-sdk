.PHONY: all test get_deps

all: test install

NOVENDOR = go list github.com/tendermint/basecoin/... | grep -v /vendor/

build:
	go build github.com/tendermint/basecoin/cmd/...

install:
	go install github.com/tendermint/basecoin/cmd/...

test:
	go test --race `${NOVENDOR}`
	#go run tests/tendermint/*.go

get_deps:
	go get -d github.com/tendermint/basecoin/...

update_deps:
	go get -d -u github.com/tendermint/basecoin/...

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install

