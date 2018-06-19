PACKAGES=$(shell go list ./... | grep -v '/vendor/')
PACKAGES_NOCLITEST=$(shell go list ./... | grep -v '/vendor/' | grep -v github.com/cosmos/cosmos-sdk/cmd/gaia/cli_test)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -tags netgo -ldflags "-X github.com/cosmos/cosmos-sdk/version.GitCommit=${COMMIT_HASH}"

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

install_debug:
	go install $(BUILD_FLAGS) ./cmd/gaia/cmd/gaiadebug

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
	@goviz -i github.com/cosmos/cosmos-sdk/cmd/gaia/cmd/gaiad -d 2 | dot -Tpng -o dependency-graph.png


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

test_race:
	@go test -race $(PACKAGES_NOCLITEST)

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


########################################
### Local validator nodes using docker and docker-compose

# Build linux binary
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

build-docker-gaiadnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: localnet-stop
	@if ! [ -f build/node0/gaiad/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/gaiad:Z tendermint/gaiadnode testnet --v 4 --o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose up

# Stop testnet
localnet-stop:
	docker-compose down

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build build_examples install install_examples install_debug dist check_tools get_tools get_vendor_deps draw_deps test test_cli test_unit test_cover test_lint benchmark devdoc_init devdoc devdoc_save devdoc_update build-linux build-docker-gaiadnode localnet-start localnet-stop
