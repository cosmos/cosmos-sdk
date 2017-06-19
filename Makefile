.PHONEY: all docs test install get_vendor_deps ensure_tools codegen

GOTOOLS = \
	github.com/Masterminds/glide
REPO:=github.com/tendermint/go-crypto

docs:
	@go get github.com/davecheney/godoc2md
	godoc2md $(REPO) > README.md

all: get_vendor_deps install test

install:
	go install ./cmd/keys

test: test_unit test_cli

test_unit:
	go test `glide novendor`
	#go run tests/tendermint/*.go

test_cli: tests/shunit2
	# sudo apt-get install jq
	@./tests/keys.sh

tests/shunit2:
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O tests/shunit2

get_vendor_deps: ensure_tools
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

ensure_tools:
	go get $(GOTOOLS)

prepgen: install
	go install ./vendor/github.com/btcsuite/btcutil/base58
	go install ./vendor/github.com/stretchr/testify/assert
	go install ./vendor/github.com/stretchr/testify/require
	go install ./vendor/golang.org/x/crypto/bcrypt

codegen:
	@echo "--> regenerating all interface wrappers"
	@gen
	@echo "Done!"
