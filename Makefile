#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
export VERSION := $(shell echo $(shell git describe --tags --always --match "v*") | sed 's/^v//')
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
  build_tags += rocksdb grocksdb_clean_link
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += boltdb
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
		-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

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

#? all: Run tools build lint test vulncheck
all: tools build lint test vulncheck

# The below include contains the tools and runsim targets.
include contrib/devtools/Makefile

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_TARGETS := build install

#? build: Build simapp
build: BUILD_ARGS=-o $(BUILDDIR)/

#? build-linux-amd64: Build simapp for GOOS=linux GOARCH=amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false $(MAKE) build

#? build-linux-arm64: Build simapp for GOOS=linux GOARCH=arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 LEDGER_ENABLED=false $(MAKE) build

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	cd ${CURRENT_DIR}/simapp && go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

#? cosmovisor: Build cosmovisor
cosmovisor:
	$(MAKE) -C tools/cosmovisor cosmovisor

#? confix: Build confix
confix:
	$(MAKE) -C tools/confix confix

#? hubl: Build hubl
hubl:
	$(MAKE) -C tools/hubl hubl

.PHONY: build build-linux-amd64 build-linux-arm64 cosmovisor confix

#? mocks: Generate mock file
mocks: $(MOCKS_DIR)
	@go install github.com/golang/mock/mockgen@v1.6.0
	sh ./scripts/mockgen.sh
.PHONY: mocks

#? vulncheck: Run govulncheck
vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

#? distclean: Run `make clean` and `make tools-clean`
distclean: clean tools-clean

#? clean: Clean some auto generated directory
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

#? godocs: Generate go doc
godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/cosmos-sdk/types"
	go install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

# make init-simapp initializes a single local node network
# it is useful for testing and development
# Usage: make install && make init-simapp && simd start
# Warning: make init-simapp will remove all data in simapp home directory
#? init-simapp: Initializes a single local node network
init-simapp:
	./scripts/init-simapp.sh

#? test: Run `make test-unit`
test: test-unit
#? test-e2e: Run `make -C tests test-e2e`
test-e2e:
	$(MAKE) -C tests test-e2e
#? test-e2e-cov: Run `make -C tests test-e2e-cov`
test-e2e-cov:
	$(MAKE) -C tests test-e2e-cov
#? test-integration: Run `make -C tests test-integration`
test-integration:
	$(MAKE) -C tests test-integration
#? test-integration-cov: Run `make -C tests test-integration-cov`
test-integration-cov:
	$(MAKE) -C tests test-integration-cov
#? test-all: Run all test
test-all: test-unit test-e2e test-integration test-ledger-mock test-race

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-amino test-unit-proto test-ledger-mock test-race test-ledger test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: test_tags += cgo ledger test_ledger_mock norace
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
$(CHECK_TEST_TARGETS): EXTRA_ARGS=-run=none
$(CHECK_TEST_TARGETS): run-tests

ARGS += -tags "$(test_tags)"
SUB_MODULES = $(shell find . -type f -name 'go.mod' -print0 | xargs -0 -n1 dirname | sort)
CURRENT_DIR = $(shell pwd)
#? run-tests: Run every sub modules' tests
run-tests:
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

#? test-sim-nondeterminism: Run non-determinism test for simapp
test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

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
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h -EnableStreaming=true

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.simapp/config/genesis.json will be used."
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -run TestFullAppSimulation -Genesis=${HOME}/.simapp/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -SigverifyTx=false -v -timeout 24h

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@cd ${CURRENT_DIR}/simapp && $(BINDIR)/runsim -Jobs=4 -SimAppPkg=. -ExitOnFail 50 5 TestAppImportExport

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	@cd ${CURRENT_DIR}/simapp && $(BINDIR)/runsim -Jobs=4 -SimAppPkg=. -ExitOnFail 50 5 TestAppSimulationAfterImport

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.simapp/config/genesis.json will be used."
	@cd ${CURRENT_DIR}/simapp && $(BINDIR)/runsim -Genesis=${HOME}/.simapp/config/genesis.json -SigverifyTx=false -SimAppPkg=. -ExitOnFail 400 5 TestFullAppSimulation

test-sim-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation. This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && $(BINDIR)/runsim -Jobs=4 -SimAppPkg=. -ExitOnFail 500 50 TestFullAppSimulation

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && $(BINDIR)/runsim -Jobs=4 -SimAppPkg=. -ExitOnFail 50 10 TestFullAppSimulation

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	cd ${CURRENT_DIR}/simapp && @go test -mod=readonly -benchmem -bench=BenchmarkInvariants -run=^$ \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 \
	-Period=1 -Commit=true -Seed=57 -v -timeout 24h

.PHONY: \
test-sim-nondeterminism \
test-sim-nondeterminism-streaming \
test-sim-custom-genesis-fast \
test-sim-import-export \
test-sim-after-import \
test-sim-custom-genesis-multi-seed \
test-sim-multi-seed-short \
test-sim-multi-seed-long \
test-sim-benchmark-invariants

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

#? test-sim-benchmark: Run benchmark test for simapp
test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

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
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -EnableStreaming=true

test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -benchmem -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

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
	@cd ${CURRENT_DIR}/simapp && go test -mod=readonly -benchmem -run=^$$ $(.) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out -EnableStreaming=true

.PHONY: test-sim-profile test-sim-benchmark

#? benchmark: Run benchmark tests
benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)
.PHONY: benchmark

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_version=v1.56.2

#? setup-pre-commit: Set pre-commit git hook
setup-pre-commit:
	@cp .git/hooks/pre-commit .git/hooks/pre-commit.bak 2>/dev/null || true
	@echo "Installing pre-commit hook..."
	@ln -sf ../../scripts/hooks/pre-commit.sh .git/hooks/pre-commit
	@echo "Pre-commit hook installed successfully"

#? lint-install: Install golangci-lint
lint-install:
	@echo "--> Installing golangci-lint $(golangci_version)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)

#? lint: Run golangci-lint
lint:
	@echo "--> Running linter"
	$(MAKE) lint-install
	@./scripts/go-lint-all.bash --timeout=15m

#? lint: Run golangci-lint and fix
lint-fix:
	@echo "--> Running linter"
	$(MAKE) lint-install
	@./scripts/go-lint-all.bash --fix

.PHONY: lint lint-fix

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

#? proto-all: Run make proto-format proto-lint proto-gen
proto-all: proto-format proto-lint proto-gen

#? proto-gen: Generate Protobuf files
proto-gen:
	@$(protoImage) sh ./scripts/protocgen.sh

#? proto-swagger-gen: Generate Protobuf Swagger
proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./scripts/protoc-swagger-gen.sh

#? proto-format: Format proto file
proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

#? proto-lint: Lint proto file
proto-lint:
	@$(protoImage) buf lint --error-format=json

#? proto-check-breaking: Check proto file is breaking
proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

#? proto-update-deps: Update protobuf dependencies
proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps

###############################################################################
###                                Localnet                                 ###
###############################################################################

#? localnet-build-env: Run `make -C contrib/images simd-env`
localnet-build-env:
	$(MAKE) -C contrib/images simd-env
#? localnet-build-dlv: Run `make -C contrib/images simd-dlv`
localnet-build-dlv:
	$(MAKE) -C contrib/images simd-dlv
#? localnet-build-nodes: Start localnet node
localnet-build-nodes:
	$(DOCKER) run --rm -v $(CURDIR)/.testnets:/data cosmossdk/simd \
			  testnet init-files -n 4 -o /data --starting-ip-address 192.168.10.2 --keyring-backend=test --listen-ip-address 0.0.0.0
	docker-compose up -d

#? localnet-stop: Stop localnet node
localnet-stop:
	docker-compose down

# localnet-start will run a 4-node testnet locally. The nodes are
# based off the docker images in: ./contrib/images/simd-env
#? localnet-start: Run a 4-node testnet locally
localnet-start: localnet-stop localnet-build-env localnet-build-nodes

# localnet-debug will run a 4-node testnet locally in debug mode
# you can read more about the debug mode here: ./contrib/images/simd-dlv/README.md
#? localnet-debug: Run a 4-node testnet locally in debug mode
localnet-debug: localnet-stop localnet-build-dlv localnet-build-nodes

.PHONY: localnet-start localnet-stop localnet-debug localnet-build-env localnet-build-dlv localnet-build-nodes

#? help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
.PHONY: help