#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')

# Ensure all tags are fetched
VERSION_RAW := $(shell git fetch --tags --force >/dev/null 2>&1; git describe --tags --always --match "v*")
VERSION := $(shell echo $(VERSION_RAW) | sed -E 's/^v?([0-9]+\.[0-9]+\.[0-9]+.*)/\1/')

# Fallback if the version is just a commit hash (not semver-like)
ifeq ($(findstring -,$(VERSION)),)  # No "-" means it's just a hash
    VERSION := 0.0.0-$(VERSION_RAW)
endif
export VERSION
export CMTVERSION := $(shell go list -m github.com/cometbft/cometbft/v2 | sed 's:.* ::')
export COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
SIMAPP = ./simapp
MOCKS_DIR = $(CURDIR)/tests/mocks
HTTPS_GIT := https://github.com/cosmos/cosmos-sdk.git
DOCKER := $(shell which docker)
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
	GCCEXE = $(shell where gcc.exe 2> NUL)
	ifeq ($(GCCEXE),)
		$(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
	else
		build_tags += ledger
	endif
	else
	UNAME_S = $(shell uname -s)
	ifeq ($(UNAME_S),OpenBSD)
		$(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
	else
		GCC = $(shell command -v gcc 2> /dev/null)
		ifeq ($(GCC),)
			$(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
		else
			build_tags += ledger
		endif
	endif
	endif
endif

ifeq (secp,$(findstring secp,$(COSMOS_BUILD_OPTIONS)))
  build_tags += libsecp256k1_sdk
endif

ifeq (legacy,$(findstring legacy,$(COSMOS_BUILD_OPTIONS)))
  build_tags += app_v1
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=sim \
		-X github.com/cosmos/cosmos-sdk/version.AppName=simd \
		-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		-X github.com/cometbft/cometbft/v2/version.TMCoreSemVer=$(CMTVERSION)

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += boltdb
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Check for debug option
ifeq (debug,$(findstring debug,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

all: tools build lint test vulncheck

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

build-linux-amd64:
	GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false $(MAKE) build

build-linux-arm64:
	GOOS=linux GOARCH=arm64 LEDGER_ENABLED=false $(MAKE) build

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	cd ${CURRENT_DIR}/simapp && go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

cosmovisor:
	$(MAKE) -C tools/cosmovisor cosmovisor

confix:
	$(MAKE) -C tools/confix confix

hubl:
	$(MAKE) -C tools/hubl hubl

.PHONY: build build-linux-amd64 build-linux-arm64 cosmovisor confix


#? mocks: Generate mock file
mocks: $(MOCKS_DIR)
	@go install go.uber.org/mock/mockgen@v0.5.0
	sh ./scripts/mockgen.sh
.PHONY: mocks


vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

distclean: clean tools-clean
clean:
	rm -rf \
	$(BUILDDIR)/ \
	artifacts/ \
	tmp-swagger-gen/ \
	.testnets

.PHONY: distclean clean

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy

###############################################################################
###                              Documentation                              ###
###############################################################################

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/cosmos-sdk/types"
	go install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060

build-docs:
	@cd docs && DOCS_DOMAIN=docs.cosmos.network sh ./build-all.sh

.PHONY: build-docs

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

# make init-simapp initializes a single local node network
# it is useful for testing and development
# Usage: make install && make init-simapp && simd start
# Warning: make init-simapp will remove all data in simapp home directory
init-simapp:
	./scripts/init-simapp.sh

test: test-unit
test-e2e:
	$(MAKE) -C tests test-e2e
test-e2e-cov:
	$(MAKE) -C tests test-e2e-cov
test-integration:
	$(MAKE) -C tests test-integration
test-integration-cov:
	$(MAKE) -C tests test-integration-cov
test-all: test-unit test-e2e test-integration test-ledger-mock test-race

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-amino test-unit-proto test-ledger-mock test-race test-ledger test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: test_tags += cgo ledger test_ledger_mock norace
test-unit-amino: test_tags += ledger test_ledger_mock test_amino norace
test-ledger: test_tags += cgo ledger norace
test-ledger-mock: test_tags += ledger test_ledger_mock norace
test-race: test_tags += cgo ledger test_ledger_mock
test-race: ARGS=-race
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

# check-* compiles and collects tests without running them
# note: go test -c doesn't support multiple packages yet (https://github.com/golang/go/issues/15513)
CHECK_TEST_TARGETS := check-test-unit check-test-unit-amino
check-test-unit: test_tags += cgo ledger test_ledger_mock norace
check-test-unit-amino: test_tags += ledger test_ledger_mock test_amino norace
$(CHECK_TEST_TARGETS): EXTRA_ARGS=-run=none
$(CHECK_TEST_TARGETS): run-tests

ARGS += -tags "$(test_tags)"
SUB_MODULES = $(shell find . -type f -name 'go.mod' -print0 | xargs -0 -n1 dirname | sort | grep -v './tests/systemtests')
CURRENT_DIR = $(shell pwd)
run-tests:
	@(cd store/streaming/abci/examples/file && go build .)
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "Starting unit tests"; \
	finalec=0; \
	for module in $(SUB_MODULES); do \
		cd ${CURRENT_DIR}/$$module; \
		echo "Running unit tests for $$(grep '^module' go.mod)"; \
		go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) ./... | tparse; \
		ec=$$?; \
		if [ "$$ec" -ne '0' ]; then finalec=$$ec; fi; \
	done; \
	exit $$finalec
else
	@echo "Starting unit tests"; \
	finalec=0; \
	for module in $(SUB_MODULES); do \
		cd ${CURRENT_DIR}/$$module; \
		echo "Running unit tests for $$(grep '^module' go.mod)"; \
		go test -mod=readonly $(ARGS) $(TEST_PACKAGES) ./... ; \
		ec=$$?; \
		if [ "$$ec" -ne '0' ]; then finalec=$$ec; fi; \
	done; \
	exit $$finalec
endif

.PHONY: run-tests test test-all $(TEST_TARGETS)

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout=30m -tags='sims' -run TestAppStateDeterminism \
		-NumBlocks=100 -BlockSize=200 -Period=0

# Requires an exported plugin. See store/streaming/README.md for documentation.
#
# example:
#   export COSMOS_SDK_ABCI_V1=<path-to-plugin-binary>
#   make test-sim-nondeterminism-streaming
#
# Using the built-in examples:
#   export COSMOS_SDK_ABCI_V1=<path-to-sdk>/store/streaming/abci/examples/file/file
#   make test-sim-nondeterminism-streaming
test-sim-nondeterminism-streaming:
	@echo "Running non-determinism-streaming test..."
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout=30m -tags='sims' -run TestAppStateDeterminism \
		-NumBlocks=100 -BlockSize=200 -Period=0 -EnableStreaming=true

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.simapp/config/genesis.json will be used."
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout=30m -tags='sims' -run TestFullAppSimulation -Genesis=${HOME}/.simapp/config/genesis.json \
		-NumBlocks=100 -BlockSize=200 -Seed=99 -Period=5 -SigverifyTx=false

test-sim-import-export:
	@echo "Running application import/export simulation. This may take several minutes..."
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout 20m -tags='sims' -run TestAppImportExport \
		-NumBlocks=50 -Period=5

test-sim-after-import:
	@echo "Running application simulation-after-import. This may take several minutes..."
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout 30m -tags='sims' -run TestAppSimulationAfterImport \
		-NumBlocks=50 -Period=5

test-sim-custom-genesis-multi-seed:
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.simapp/config/genesis.json will be used."
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout 30m -tags='sims' -run TestFullAppSimulation -Genesis=${HOME}/.simapp/config/genesis.json \
		-NumBlocks=400 -Period=5

test-sim-multi-seed-long:
	@echo "Running long multi-seed application simulation. This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout=1h -tags='sims' -run TestFullAppSimulation \
		-NumBlocks=500 -Period=50

test-sim-multi-seed-short:
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -timeout 30m -tags='sims' -run TestFullAppSimulation \
		-NumBlocks=50 -Period=10

.PHONY: \
test-sim-nondeterminism \
test-sim-nondeterminism-streaming \
test-sim-custom-genesis-fast \
test-sim-import-export \
test-sim-after-import \
test-sim-custom-genesis-multi-seed \
test-sim-multi-seed-short \
test-sim-multi-seed-long

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

#? test-sim-fuzz: Run fuzz test for simapp
test-sim-fuzz:
	@echo "Running application fuzz for numBlocks=2, blockSize=20. This may take awhile!"
#ld flags are a quick fix to make it work on current osx
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -json -tags='sims' -ldflags="-extldflags=-Wl,-ld_classic" -timeout=60m -fuzztime=60m -run=^$$ -fuzz=FuzzFullAppSimulation -GenesisTime=1714720615 -NumBlocks=2 -BlockSize=20

#? test-sim-benchmark: Run benchmark test for simapp
test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -tags='sims' -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$  \
		-NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -Seed=57 -timeout 30m

# Requires an exported plugin. See store/streaming/README.md for documentation.
#
# example:
#   export COSMOS_SDK_ABCI_V1=<path-to-plugin-binary>
#   make test-sim-benchmark-streaming
#
# Using the built-in examples:
#   export COSMOS_SDK_ABCI_V1=<path-to-sdk>/store/streaming/abci/examples/file/file
#   make test-sim-benchmark-streaming
test-sim-benchmark-streaming:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$  \
		-NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -EnableStreaming=true

test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -benchmem -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$ \
		-NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

# Requires an exported plugin. See store/streaming/README.md for documentation.
#
# example:
#   export COSMOS_SDK_ABCI_V1=<path-to-plugin-binary>
#   make test-sim-profile-streaming
#
# Using the built-in examples:
#   export COSMOS_SDK_ABCI_V1=<path-to-sdk>/store/streaming/abci/examples/file/file
#   make test-sim-profile-streaming
test-sim-profile-streaming:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -failfast -mod=readonly -benchmem -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$ \
		-NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out -EnableStreaming=true

.PHONY: test-sim-profile test-sim-benchmark test-sim-fuzz

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)
.PHONY: benchmark

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_version=v2.1.6

lint-install:
	@echo "--> Installing golangci-lint $(golangci_version)"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)

lint:
	@echo "--> Running linter on all files"
	$(MAKE) lint-install
	@./scripts/go-lint-all.bash --timeout=15m


lint-fix:
	@echo "--> Running linter"
	$(MAKE) lint-install
	@./scripts/go-lint-all.bash --fix

.PHONY: lint lint-fix

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.17.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh 2>&1 | tee protocgen.log | \
	awk '{print $$0} /contains the reserved field name/ && /tendermint/ {next} 1'


proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./scripts/protoc-swagger-gen.sh

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

CMT_VERSION_DIR      = v1.0.1
CMT_PROTO            = v1
CMT_URL              = https://raw.githubusercontent.com/cometbft/cometbft/$(CMT_VERSION_DIR)/proto/cometbft
CMT_CRYPTO_TYPES     = proto/cometbft/crypto/$(CMT_PROTO)
CMT_ABCI_TYPES       = proto/cometbft/abci/$(CMT_PROTO)
CMT_TYPES            = proto/cometbft/types/$(CMT_PROTO)
CMT_VERSION          = proto/cometbft/version/$(CMT_PROTO)
CMT_LIBS             = proto/cometbft/libs/bits/$(CMT_PROTO)
CMT_P2P              = proto/cometbft/p2p/$(CMT_PROTO)

proto-update-comet:
	@echo "Updating Protobuf dependency: downloading cometbft.$(CMT_PROTO) files from CometBFT $(CMT_VERSION_DIR)"

	@mkdir -p $(CMT_ABCI_TYPES)
	@curl -fsSL $(CMT_URL)/abci/$(CMT_PROTO)/service.proto > $(CMT_ABCI_TYPES)/service.proto
	@curl -fsSL $(CMT_URL)/abci/$(CMT_PROTO)/types.proto > $(CMT_ABCI_TYPES)/types.proto

	@mkdir -p $(CMT_VERSION)
	@curl -fsSL $(CMT_URL)/version/$(CMT_PROTO)/types.proto > $(CMT_VERSION)/types.proto

	@mkdir -p $(CMT_TYPES)
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/block.proto > $(CMT_TYPES)/block.proto
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/canonical.proto > $(CMT_TYPES)/canonical.proto
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/events.proto > $(CMT_TYPES)/events.proto
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/evidence.proto > $(CMT_TYPES)/evidence.proto
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/params.proto > $(CMT_TYPES)/params.proto
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/types.proto > $(CMT_TYPES)/types.proto
	@curl -fsSL $(CMT_URL)/types/$(CMT_PROTO)/validator.proto > $(CMT_TYPES)/validator.proto

	@mkdir -p $(CMT_CRYPTO_TYPES)
	@curl -fsSL $(CMT_URL)/crypto/$(CMT_PROTO)/keys.proto > $(CMT_CRYPTO_TYPES)/keys.proto
	@curl -fsSL $(CMT_URL)/crypto/$(CMT_PROTO)/proof.proto > $(CMT_CRYPTO_TYPES)/proof.proto

	@mkdir -p $(CMT_LIBS)
	@curl -fsSL $(CMT_URL)/libs/bits/$(CMT_PROTO)/types.proto > $(CMT_LIBS)/types.proto

	@mkdir -p $(CMT_P2P)
	@curl -fsSL $(CMT_URL)/p2p/$(CMT_PROTO)/conn.proto > $(CMT_P2P)/conn.proto
	@curl -fsSL $(CMT_URL)/p2p/$(CMT_PROTO)/pex.proto > $(CMT_P2P)/pex.proto
	@curl -fsSL $(CMT_URL)/p2p/$(CMT_PROTO)/types.proto > $(CMT_P2P)/types.proto

proto-update-deps:
	@echo "Updating Protobuf dependencies: running 'buf dep update'"
	$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf dep update

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps proto-update-comet

###############################################################################
###                                Localnet                                 ###
###############################################################################

localnet-build-env:
	$(MAKE) -C contrib/images simd-env
localnet-build-dlv:
	$(MAKE) -C contrib/images simd-dlv

localnet-build-nodes:
	$(DOCKER) run --rm -v $(CURDIR)/.testnets:/data cosmossdk/simd \
			  testnet init-files --validator-count 4 -o /data --starting-ip-address 192.168.10.2 --keyring-backend=test
	docker compose up -d

localnet-stop:
	docker compose down

# localnet-start will run a 4-node testnet locally. The nodes are
# based off the docker images in: ./contrib/images/simd-env
localnet-start: localnet-stop localnet-build-env localnet-build-nodes

# localnet-debug will run a 4-node testnet locally in debug mode
# you can read more about the debug mode here: ./contrib/images/simd-dlv/README.md
localnet-debug: localnet-stop localnet-build-dlv localnet-build-nodes

.PHONY: localnet-start localnet-stop localnet-debug localnet-build-env localnet-build-dlv localnet-build-nodes

test-system: build-v53 build
	mkdir -p ./tests/systemtests/binaries/
	cp $(BUILDDIR)/simd ./tests/systemtests/binaries/
	mkdir -p ./tests/systemtests/binaries/v0.53
	mv $(BUILDDIR)/simdv53 ./tests/systemtests/binaries/v0.53/simd
	$(MAKE) -C tests/systemtests test
.PHONY: test-system

# build-v53 checks out the v0.53.x branch, builds the binary, and renames it to simdv53.
build-v53:
	@echo "Starting v53 build process..."
	git_status=$$(git status --porcelain) && \
	has_changes=false && \
	if [ -n "$$git_status" ]; then \
		echo "Stashing uncommitted changes..." && \
		git stash push -m "Temporary stash for v53 build" && \
		has_changes=true; \
	else \
		echo "No changes to stash"; \
	fi && \
	echo "Saving current reference..." && \
	CURRENT_REF=$$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse HEAD) && \
	echo "Checking out release branch..." && \
	git checkout release/v0.53.x && \
	echo "Building v53 binary..." && \
	make build && \
	mv build/simd build/simdv53 && \
	echo "Returning to original branch..." && \
	if [ "$$CURRENT_REF" = "HEAD" ]; then \
		git checkout $$(git rev-parse HEAD); \
	else \
		git checkout $$CURRENT_REF; \
	fi && \
	if [ "$$has_changes" = "true" ]; then \
		echo "Reapplying stashed changes..." && \
		git stash pop || echo "Warning: Could not pop stash, your changes may be in the stash list"; \
	else \
		echo "No changes to reapply"; \
	fi
.PHONY: build-v53