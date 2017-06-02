GOTOOLS = \
					github.com/mitchellh/gox \
					github.com/Masterminds/glide
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: test install

build:
	go build ./cmd/...

install:
	go install ./cmd/...

dist:
	@bash scripts/dist.sh
	@bash scripts/publish.sh

test:
	go test $(PACKAGES)
	#go run tests/tendermint/*.go

get_deps:
	go get -d ./...

update_deps:
	go get -d -u ./...

get_vendor_deps: tools
	glide install

build-docker:
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/basecoin" -w "/go/src/github.com/tendermint/basecoin" -e "CGO_ENABLED=0" golang:alpine go build ./cmd/basecoin
	docker build -t "tendermint/basecoin" .

tools:
	go get -u -v $(GOTOOLS)

clean:
	@rm -f ./basecoin

.PHONY: all build install test get_deps update_deps get_vendor_deps build-docker clean
