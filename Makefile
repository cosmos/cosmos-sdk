.PHONEY: all test install get_vendor_deps ensure_tools codegen wordlist

GOTOOLS = \
	github.com/Masterminds/glide \
	github.com/jteeuwen/go-bindata/go-bindata
REPO:=github.com/tendermint/go-crypto

all: get_vendor_deps test

test:
	go test `glide novendor`

get_vendor_deps: ensure_tools
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

ensure_tools:
	go get $(GOTOOLS)

wordlist:
	go-bindata -ignore ".*\.go" -o keys/wordlist/wordlist.go -pkg "wordlist" keys/wordlist/...

prepgen: install
	go install ./vendor/github.com/btcsuite/btcutil/base58
	go install ./vendor/github.com/stretchr/testify/assert
	go install ./vendor/github.com/stretchr/testify/require
	go install ./vendor/golang.org/x/crypto/bcrypt

codegen:
	@echo "--> regenerating all interface wrappers"
	@gen
	@echo "Done!"
