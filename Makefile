.PHONEY: all docs test install get_vendor_deps ensure_tools

GOTOOLS = \
	github.com/Masterminds/glide
REPO:=github.com/tendermint/go-crypto

docs:
	@go get github.com/davecheney/godoc2md
	godoc2md $(REPO) > README.md

all: install test

install:
	go install ./cmd/keys

test:
	go test `glide novendor`

get_vendor_deps: ensure_tools
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

ensure_tools:
	go get $(GOTOOLS)


