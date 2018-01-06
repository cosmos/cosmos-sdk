GOTOOLS =	github.com/mitchellh/gox \
			github.com/Masterminds/glide \
			github.com/rigelrozanski/shelldown/cmd/shelldown
TUTORIALS=$(shell find docs/guide -name "*md" -type f)

EXAMPLES := dummy basecoin

INSTALL_EXAMPLES := $(addprefix install_,${EXAMPLES})
TEST_EXAMPLES := $(addprefix testex_,${EXAMPLES})

LINKER_FLAGS:="-X github.com/cosmos/cosmos-sdk/client/commands.CommitHash=`git rev-parse --short HEAD`"

all: get_vendor_deps install test

$(INSTALL_EXAMPLES): install_%:
	cd ./examples/$* && go install

$(TEST_EXAMPLES): testex_%:
	cd ./examples/$* && make test_cli

install: $(INSTALL_EXAMPLES)

dist:
	@bash publish/dist.sh
	@bash publish/publish.sh

benchmark:
	@go test -bench=. ./modules/...

#test: test_unit test_cli test_tutorial
test: test_unit # test_cli

test_unit:
	@go test `glide novendor | grep -v _attic`

test_cli: $(TEST_EXAMPLES)
	# sudo apt-get install jq
	# wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2"

test_tutorial:
	@shelldown ${TUTORIALS}
	@for script in docs/guide/*.sh ; do \
		bash $$script ; \
	done

get_vendor_deps: get_tools
	@glide install

build-docker:
	@docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/basecoin" -w \
		"/go/src/github.com/tendermint/basecoin" -e "CGO_ENABLED=0" golang:alpine go build ./cmd/basecoin
	@docker build -t "tendermint/basecoin" .

get_tools:
	@go get $(GOTOOLS)

clean:
	# maybe cleaning up cache and vendor is overkill, but sometimes
	# you don't get the most recent versions with lots of branches, changes, rebases...
	@rm -rf ~/.glide/cache/src/https-github.com-tendermint-*
	@rm -rf ./vendor
	@rm -f $GOPATH/bin/{basecoin,basecli,counter,countercli}

# when your repo is getting a little stale... just make fresh
fresh: clean get_vendor_deps install
	@if [ "$(git status -s)" ]; then echo; echo "Warning: uncommited changes"; git status -s; fi

.PHONY: all build install test test_cli test_unit test_store get_vendor_deps build-docker clean fresh benchmark
