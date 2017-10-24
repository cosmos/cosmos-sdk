.PHONEY: all test install get_vendor_deps ensure_tools codegen wordlist

GOTOOLS = \
	github.com/Masterminds/glide \
	github.com/jteeuwen/go-bindata/go-bindata \
	github.com/alecthomas/gometalinter

REPO:=github.com/tendermint/go-crypto

all: get_vendor_deps metalinter_test test

test:
	go test -p 1 `glide novendor`

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

metalinter: ensure_tools
	@gometalinter --install
	gometalinter --vendor --deadline=600s --enable-all --disable=lll ./...

metalinter_test: ensure_tools
	@gometalinter --install
	gometalinter --vendor --deadline=600s --disable-all  \
		--enable=deadcode \
		--enable=gas \
		--enable=goconst \
		--enable=gocyclo \
		--enable=gosimple \
	 	--enable=ineffassign \
	   	--enable=interfacer \
		--enable=maligned \
		--enable=megacheck \
	 	--enable=misspell \
		--enable=safesql \
		--enable=staticcheck \
		--enable=structcheck \
	   	--enable=unconvert \
		--enable=unused \
		--enable=vetshadow \
		--enable=vet \
		--enable=varcheck \
		./...

		#--enable=dupl \
		#--enable=errcheck \
		#--enable=goimports \
		#--enable=golint \ <== comments on anything exported
		#--enable=gotype \
		#--enable=unparam \
