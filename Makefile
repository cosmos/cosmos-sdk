.PHONEY: all docs test install get_vendor_deps ensure_tools codegen

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

prepgen: install
	cd ../go-wire && make tools
	go install ./vendor/github.com/btcsuite/btcutil/base58
	go install ./vendor/github.com/stretchr/testify/assert
	go install ./vendor/github.com/stretchr/testify/require
	go install ./vendor/golang.org/x/crypto/bcrypt

codegen: prepgen
	gen
