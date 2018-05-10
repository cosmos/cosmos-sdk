PACKAGES=$(shell go list ./... | grep -v '/vendor/')
PACKAGES_NOCLITEST=$(shell go list ./... | grep -v '/vendor/' | grep -v github.com/cosmos/cosmos-sdk/cmd/gaia/cli_test)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -ldflags "-X github.com/cosmos/cosmos-sdk/version.GitCommit=${COMMIT_HASH}"

all: check_tools get_vendor_deps install install_examples test_lint test

########################################
### CI

ci: get_tools get_vendor_deps install test_cover test_lint test

########################################
### Build

# This can be unified later, here for easy demos
build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/gaiad.exe ./cmd/gaia/cmd/gaiad
	go build $(BUILD_FLAGS) -o build/gaiacli.exe ./cmd/gaia/cmd/gaiacli
else
	go build $(BUILD_FLAGS) -o build/gaiad ./cmd/gaia/cmd/gaiad
	go build $(BUILD_FLAGS) -o build/gaiacli ./cmd/gaia/cmd/gaiacli
endif

build_examples:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/basecoind.exe ./examples/basecoin/cmd/basecoind
	go build $(BUILD_FLAGS) -o build/basecli.exe ./examples/basecoin/cmd/basecli
	go build $(BUILD_FLAGS) -o build/democoind.exe ./examples/democoin/cmd/democoind
	go build $(BUILD_FLAGS) -o build/democli.exe ./examples/democoin/cmd/democli
else
	go build $(BUILD_FLAGS) -o build/basecoind ./examples/basecoin/cmd/basecoind
	go build $(BUILD_FLAGS) -o build/basecli ./examples/basecoin/cmd/basecli
	go build $(BUILD_FLAGS) -o build/democoind ./examples/democoin/cmd/democoind
	go build $(BUILD_FLAGS) -o build/democli ./examples/democoin/cmd/democli
endif

install: 
	go install $(BUILD_FLAGS) ./cmd/gaia/cmd/gaiad
	go install $(BUILD_FLAGS) ./cmd/gaia/cmd/gaiacli

install_examples: 
	go install $(BUILD_FLAGS) ./examples/basecoin/cmd/basecoind
	go install $(BUILD_FLAGS) ./examples/basecoin/cmd/basecli
	go install $(BUILD_FLAGS) ./examples/democoin/cmd/democoind
	go install $(BUILD_FLAGS) ./examples/democoin/cmd/democli

dist:
	@bash publish/dist.sh
	@bash publish/publish.sh

########################################
### Tools & dependencies

check_tools:
	cd tools && $(MAKE) check_tools

update_tools:
	cd tools && $(MAKE) update_tools

get_tools:
	cd tools && $(MAKE) get_tools

get_vendor_deps:
	@rm -rf vendor/
	@echo "--> Running dep ensure"
	@dep ensure -v

draw_deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i github.com/tendermint/tendermint/cmd/tendermint -d 3 | dot -Tpng -o dependency-graph.png


########################################
### Documentation

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/cosmos-sdk/types"
	godoc -http=:6060


########################################
### Testing

test: test_unit

test_cli: 
	@go test -count 1 -p 1 `go list github.com/cosmos/cosmos-sdk/cmd/gaia/cli_test`

test_unit:
	@go test $(PACKAGES_NOCLITEST)

test_cover:
	@bash tests/test_cover.sh

test_lint:
	gometalinter --disable-all --enable='golint' --vendor ./...

benchmark:
	@go test -bench=. $(PACKAGES_NOCLITEST)


########################################
### Devdoc

DEVDOC_SAVE = docker commit `docker ps -a -n 1 -q` devdoc:local

devdoc_init:
	docker run -it -v "$(CURDIR):/go/src/github.com/cosmos/cosmos-sdk" -w "/go/src/github.com/cosmos/cosmos-sdk" tendermint/devdoc echo
	# TODO make this safer
	$(call DEVDOC_SAVE)

devdoc:
	docker run -it -v "$(CURDIR):/go/src/github.com/cosmos/cosmos-sdk" -w "/go/src/github.com/cosmos/cosmos-sdk" devdoc:local bash

devdoc_save:
	# TODO make this safer
	$(call DEVDOC_SAVE)

devdoc_clean:
	docker rmi -f $$(docker images -f "dangling=true" -q)

devdoc_update:
	docker pull tendermint/devdoc


# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build build_examples install install_examples dist check_tools get_tools get_vendor_deps draw_deps test test_cli test_unit test_cover test_lint benchmark devdoc_init devdoc devdoc_save devdoc_update
