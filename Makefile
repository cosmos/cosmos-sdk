GOTOOLS = \
	github.com/mitchellh/gox \
	github.com/Masterminds/glide

.PHONEY: all test install get_vendor_deps ensure_tools

all: install test

test:
	go test `glide novendor` 

install:
	go install ./cmd/keys

get_vendor_deps: ensure_tools
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

ensure_tools:
	go get $(GOTOOLS)

