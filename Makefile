#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
SIMAPP = ./simapp
MOCKS_DIR = $(CURDIR)/tests/mocks
HTTPS_GIT := https://github.com/cosmos/cosmos-sdk.git
DOCKER_BUF := docker run -v $(shell pwd):/workspace --workdir /workspace bufbuild/buf

export GO111MODULE = on

# The below include contains the tools and runsim targets.
include contrib/devtools/Makefile

all: tools build lint test

###############################################################################
###                                  Build                                  ###
###############################################################################

build: go.sum
	@go build -mod=readonly ./...

build-sim: go.sum
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR) ./simapp/cmd/simd
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR) ./simapp/cmd/simcli

.PHONY: \
 build \
 build-sim

mocks: $(MOCKS_DIR)
	mockgen -source=client/context/account_retriever.go -package mocks -destination tests/mocks/account_retriever.go
	mockgen -package mocks -destination tests/mocks/tendermint_tm_db_DB.go github.com/tendermint/tm-db DB
	mockgen -source=types/module/module.go -package mocks -destination tests/mocks/types_module_module.go
	mockgen -source=types/invariant.go -package mocks -destination tests/mocks/types_invariant.go
	mockgen -source=types/router.go -package mocks -destination tests/mocks/types_router.go
	mockgen -source=types/handler.go -package mocks -destination tests/mocks/types_handler.go
.PHONY: mocks

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

distclean:
	rm -rf \
    gitian-build-darwin/ \
    gitian-build-linux/ \
    gitian-build-windows/ \
    .gitian-builder-cache/
.PHONY: distclean

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
.PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

###############################################################################
###                              Documentation                              ###
###############################################################################

update-swagger-docs: statik
	$(BINDIR)/statik -src=client/lcd/swagger-ui -dest=client/lcd -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
    	echo "\033[92mSwagger docs are in sync\033[0m";\
    fi
.PHONY: update-swagger-docs

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/cosmos-sdk/types"
	godoc -http=:6060

build-docs:
	@cd docs && \
	while read p; do \
		(git checkout $${p} && npm install && VUEPRESS_BASE="/$${p}/" npm run build) ; \
		mkdir -p ~/output/$${p} ; \
		cp -r .vuepress/dist/* ~/output/$${p}/ ; \
		cp ~/output/$${p}/index.html ~/output ; \
	done < versions ;

sync-docs:
	cd ~/output && \
	echo "role_arn = ${DEPLOYMENT_ROLE_ARN}" >> /root/.aws/config ; \
	echo "CI job = ${CIRCLE_BUILD_URL}" >> version.html ; \
	aws s3 sync . s3://${WEBSITE_BUCKET} --profile terraform --delete ; \
	aws cloudfront create-invalidation --distribution-id ${CF_DISTRIBUTION_ID} --profile terraform --path "/*" ;
.PHONY: sync-docs

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test: test-unit
test-all: test-unit test-ledger-mock test-race test-cover

test-ledger-mock:
	@go test -mod=readonly `go list github.com/cosmos/cosmos-sdk/crypto` -tags='cgo ledger test_ledger_mock'

test-ledger: test-ledger-mock
	@go test -mod=readonly -v `go list github.com/cosmos/cosmos-sdk/crypto` -tags='cgo ledger'

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly ./... -tags='ledger test_ledger_mock'

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race $(PACKAGES_NOSIMULATION)

test-integration: build-sim
	BUILDDIR=$(BUILDDIR) go test -mod=readonly -p 4 -tags='ledger test_ledger_mock cli_test' -run ^TestCLI `go list ./.../cli/...`

.PHONY: test test-all test-ledger-mock test-ledger test-unit test-race

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=${HOME}/.gaiad/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 50 5 TestAppImportExport

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 50 5 TestAppSimulationAfterImport

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@$(BINDIR)/runsim -Genesis=${HOME}/.gaiad/config/genesis.json -SimAppPkg=$(SIMAPP) 400 5 TestFullAppSimulation

test-sim-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 500 50 TestFullAppSimulation

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 50 10 TestFullAppSimulation

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkInvariants -run=^$ \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 \
	-Period=1 -Commit=true -Seed=57 -v -timeout 24h

.PHONY: \
test-sim-nondeterminism \
test-sim-custom-genesis-fast \
test-sim-import-export \
test-sim-after-import \
test-sim-custom-genesis-multi-seed \
test-sim-multi-seed-short \
test-sim-multi-seed-long \
test-sim-benchmark-invariants \
test-integration

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

.PHONY: test-sim-profile test-sim-benchmark

test-cover:
	@export VERSION=$(VERSION); bash -x contrib/test_cover.sh
.PHONY: test-cover

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)
.PHONY: benchmark

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --out-format=tab --issues-exit-code=0
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
.PHONY: lint

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" -not -path "./tests/mocks/*" -not -name '*.pb.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" -not -path "./tests/mocks/*" -not -name '*.pb.go' | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" -not -path "./tests/mocks/*" -not -name '*.pb.go' | xargs goimports -w -local github.com/cosmos/cosmos-sdk
.PHONY: format

###############################################################################
###                                 Devdoc                                  ###
###############################################################################

DEVDOC_SAVE = docker commit `docker ps -a -n 1 -q` devdoc:local

devdoc-init:
	docker run -it -v "$(CURDIR):/go/src/github.com/cosmos/cosmos-sdk" -w "/go/src/github.com/cosmos/cosmos-sdk" tendermint/devdoc echo
	# TODO make this safer
	$(call DEVDOC_SAVE)

devdoc:
	docker run -it -v "$(CURDIR):/go/src/github.com/cosmos/cosmos-sdk" -w "/go/src/github.com/cosmos/cosmos-sdk" devdoc:local bash

devdoc-save:
	# TODO make this safer
	$(call DEVDOC_SAVE)

devdoc-clean:
	docker rmi -f $$(docker images -f "dangling=true" -q)

devdoc-update:
	docker pull tendermint/devdoc

.PHONY: devdoc devdoc-clean devdoc-init devdoc-save devdoc-update

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-tools proto-gen proto-lint proto-check-breaking

proto-gen:
	@./scripts/protocgen.sh

# This generates the SDK's custom wrapper for google.protobuf.Any. It should only be run manually when needed
proto-gen-any:
	@./scripts/protocgen-any.sh

proto-lint:
	@buf check lint --error-format=json

proto-check-breaking:
	@buf check breaking --against-input '.git#branch=master'

proto-lint-docker:
	@$(DOCKER_BUF) check lint --error-format=json
.PHONY: proto-lint

proto-check-breaking-docker:
	@$(DOCKER_BUF) check breaking --against-input $(HTTPS_GIT)#branch=master
.PHONY: proto-check-breaking-ci

TM_URL           = https://raw.githubusercontent.com/tendermint/tendermint/v0.33.1
GOGO_PROTO_URL   = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_PROTO_URL = https://raw.githubusercontent.com/regen-network/cosmos-proto/master

TM_KV_TYPES         = third_party/proto/tendermint/libs/kv
TM_MERKLE_TYPES     = third_party/proto/tendermint/crypto/merkle
TM_ABCI_TYPES       = third_party/proto/tendermint/abci/types
GOGO_PROTO_TYPES    = third_party/proto/gogoproto
COSMOS_PROTO_TYPES  = third_party/proto/cosmos-proto
SDK_PROTO_TYPES     = third_party/proto/cosmos-sdk/types
AUTH_PROTO_TYPES    = third_party/proto/cosmos-sdk/x/auth/types
VESTING_PROTO_TYPES = third_party/proto/cosmos-sdk/x/auth/vesting/types
SUPPLY_PROTO_TYPES  = third_party/proto/cosmos-sdk/x/supply/types

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto

	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types/types.proto > $(TM_ABCI_TYPES)/types.proto
	@sed -i '' '8 s|crypto/merkle/merkle.proto|third_party/proto/tendermint/crypto/merkle/merkle.proto|g' $(TM_ABCI_TYPES)/types.proto
	@sed -i '' '9 s|libs/kv/types.proto|third_party/proto/tendermint/libs/kv/types.proto|g' $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_KV_TYPES)
	@curl -sSL $(TM_URL)/libs/kv/types.proto > $(TM_KV_TYPES)/types.proto

	@mkdir -p $(TM_MERKLE_TYPES)
	@curl -sSL $(TM_URL)/crypto/merkle/merkle.proto > $(TM_MERKLE_TYPES)/merkle.proto


.PHONY: proto-all proto-gen proto-lint proto-check-breaking proto-update-deps
