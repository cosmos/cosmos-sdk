GOTOOLS =	github.com/mitchellh/gox \
			github.com/Masterminds/glide \
			github.com/rigelrozanski/shelldown/cmd/shelldown
TUTORIALS=$(shell find docs/guide -name "*md" -type f)

all: get_vendor_deps install test

build:
	go build ./cmd/...

install:
	go install ./cmd/...
	go install ./docs/guide/counter/cmd/...

dist:
	@bash scripts/dist.sh
	@bash scripts/publish.sh

test: test_unit test_cli test_tutorial

test_unit:
	go test `glide novendor`
	#go run tests/tendermint/*.go

test_cli: tests/cli/shunit2
	# sudo apt-get install jq
	@./tests/cli/basictx.sh
	@./tests/cli/counter.sh
	@./tests/cli/restart.sh
	@./tests/cli/ibc.sh

test_tutorial: docs/guide/shunit2
	shelldown ${TUTORIALS}
	for script in docs/guide/*.sh ; do \
		bash $$script ; \
	done

tests/cli/shunit2:
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
    	-q -O tests/cli/shunit2

docs/guide/shunit2:
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
    	-q -O docs/guide/shunit2

get_vendor_deps: tools
	glide install

build-docker:
	docker run -it --rm -v "$(PWD):/go/src/github.com/tendermint/basecoin" -w \
		"/go/src/github.com/tendermint/basecoin" -e "CGO_ENABLED=0" golang:alpine go build ./cmd/basecoin
	docker build -t "tendermint/basecoin" .

tools:
	@go get $(GOTOOLS)

clean:
	# maybe cleaning up cache and vendor is overkill, but sometimes
	# you don't get the most recent versions with lots of branches, changes, rebases...
	@rm -rf ~/.glide/cache/src/https-github.com-tendermint-*
	@rm -rf ./vendor
	@rm -f $GOPATH/bin/{basecoin,basecli,counter,countercli}

# when your repo is getting a little stale... just make fresh
fresh: clean get_vendor_deps install
	@if [[ `git status -s` ]]; then echo; echo "Warning: uncommited changes"; git status -s; fi

.PHONY: all build install test test_cli test_unit get_vendor_deps build-docker clean fresh
