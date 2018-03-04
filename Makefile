PACKAGES=$(shell go list ./... | grep -v '/vendor/')
BUILD_FLAGS = -ldflags "-X github.com/cosmos/cosmos-sdk/version.GitCommit=`git rev-parse --short HEAD`"

all: check_tools get_vendor_deps build test

########################################
### CI

ci: get_tools get_vendor_deps build test_cover

########################################
### Build

# This can be unified later, here for easy demos
gaia:
	go build $(BUILD_FLAGS) -o build/gaiad ./examples/gaia/gaiad
	go build $(BUILD_FLAGS) -o build/gaiacli ./examples/gaia/gaiacli

build:
	@rm -rf examples/basecoin/vendor/
	go build $(BUILD_FLAGS) -o build/basecoind ./examples/basecoin/cmd/basecoind
	go build $(BUILD_FLAGS) -o build/basecli ./examples/basecoin/cmd/basecli

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

test: test_unit # test_cli

# Must  be run in each package seperately for the visualization
# Added here for easy reference
# coverage:
#	 go test -coverprofile=c.out && go tool cover -html=c.out

test_unit:
	@rm -rf examples/basecoin/vendor/
	@go test $(PACKAGES)

test_cover:
	@rm -rf examples/basecoin/vendor/
	@bash tests/test_cover.sh

benchmark:
	@go test -bench=. $(PACKAGES)


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
.PHONY: build dist check_tools get_tools get_vendor_deps draw_deps test test_unit test_tutorial benchmark devdoc_init devdoc devdoc_save devdoc_update
