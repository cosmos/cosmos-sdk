GOTOOLS =	github.com/mitchellh/gox \
			github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_vendor_deps install test

build:
	go build ./cmd/...

install:
	go install ./cmd/...

dist:
	@bash scripts/dist.sh
	@bash scripts/publish.sh

test: test_unit test_cli

test_unit:
	go test $(PACKAGES)
	#go run tests/tendermint/*.go

test_cli:
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O tests/cli/shunit2
	# sudo apt-get install jq
	@./tests/cli/basictx.sh
	# @./clitest/ibc.sh

get_vendor_deps: tools
	glide install

build-docker:
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/basecoin" -w \
		"/go/src/github.com/tendermint/basecoin" -e "CGO_ENABLED=0" golang:alpine go build ./cmd/basecoin
	docker build -t "tendermint/basecoin" .

tools:
	go get -u -v $(GOTOOLS)

clean:
	@rm -f ./basecoin

.PHONY: all build install test test_cli test_unit get_vendor_deps build-docker clean
