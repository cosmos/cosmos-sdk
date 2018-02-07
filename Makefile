all: check_tools get_vendor_deps build test

########################################
### CI

ci: get_tools get_vendor_deps build test_cover

########################################
### Build

build:
	@rm -rf examples/basecoin/vendor
	cd examples/basecoin && $(MAKE) get_vendor_deps build

dist:
	@bash publish/dist.sh
	@bash publish/publish.sh

########################################
### Tools & dependencies

check_tools:
	cd tools && $(MAKE) check

update_tools:
	cd tools && $(MAKE) glide_update

get_tools:
	cd tools && $(MAKE)

get_vendor_deps:
	@rm -rf vendor/
	@echo "--> Running glide install"
	@glide install

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

PACKAGES=$(shell go list ./... | grep -v '/vendor/' | grep -v '/examples/' | grep -v '/tools/')
TUTORIALS=$(shell find docs/guide -name "*md" -type f)

#test: test_unit test_cli test_tutorial
test: test_unit # test_cli

test_basecoin:
	@rm -rf examples/basecoin/vendor
	@cd examples/basecoin $(MAKE) get_vendor_deps test

test_unit:
	@go test $(PACKAGES)

test_cover:
	@bash test_cover.sh

test_tutorial:
	@shelldown ${TUTORIALS}
	@for script in docs/guide/*.sh ; do \
		bash $$script ; \
	done

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
.PHONY: build dist check_tools get_tools get_vendor_deps draw_deps test test_unit test_tutorial benchmark devdoc_init devdoc devdoc_save devdoc_update test_basecoin
