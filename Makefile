#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./simapp
MOCKS_DIR = $(CURDIR)/tests/mocks

export GO111MODULE = on

all: build lint test

########################################
### CI

ci: build test_cover lint test

########################################
### Build

build: go.sum
	@go build -mod=readonly ./...

update-swagger-docs:
	@$(BINDIR)/statik -src=client/lcd/swagger-ui -dest=client/lcd -f -m
	if [ -n "$(git status --porcelain)" ]; then \
		echo "swagger docs out of sync";\
        exit 1;\
    else \
    	echo "swagger docs are in sync";\
    fi

.PHONY: update-swagger-docs

mocks: $(MOCKS_DIR)
	mockgen -source=x/auth/types/account_retriever.go -package mocks -destination tests/mocks/account_retriever.go

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

distclean:
	rm -rf \
    gitian-build-darwin/ \
    gitian-build-linux/ \
    gitian-build-windows/ \
    .gitian-builder-cache/

########################################
### Documentation

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/cosmos-sdk/types"
	godoc -http=:6060


########################################
### Testing

test: test_unit

test_ledger_mock:
		@go test -mod=readonly `go list github.com/cosmos/cosmos-sdk/crypto` -tags='cgo ledger test_ledger_mock'

test_ledger: test_ledger_mock
	@go test -mod=readonly -v `go list github.com/cosmos/cosmos-sdk/crypto` -tags='cgo ledger'

test_unit:
	@VERSION=$(VERSION) go test -mod=readonly $(PACKAGES_NOSIMULATION) -tags='ledger test_ledger_mock'

test_race:
	@VERSION=$(VERSION) go test -mod=readonly -race $(PACKAGES_NOSIMULATION)

test_sim_app_nondeterminism:
	@echo "Running nondeterminism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true -v -timeout 10m

test_sim_app_custom_genesis_fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=${HOME}/.gaiad/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test_sim_app_fast:
	@echo "Running quick application simulation. This may take several minutes..."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test_sim_app_import_export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	$(BINDIR)/runsim -j 4 $(SIMAPP) 50 5 TestAppImportExport

test_sim_app_simulation_after_import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(BINDIR)/runsim -e $(SIMAPP) 25 5 TestAppSimulationAfterImport

test_sim_app_custom_genesis_multi_seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	$(BINDIR)/runsim -g ${HOME}/.gaiad/config/genesis.json $(SIMAPP) 400 5 TestFullAppSimulation

test_sim_app_multi_seed: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -j 4 $(SIMAPP) 500 50 TestFullAppSimulation

test_sim_app_multi_seed_short: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -j 4 $(SIMAPP) 50 10 TestFullAppSimulation

test_sim_benchmark_invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkInvariants -run=^$ \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 \
	-Commit=true -Seed=57 -v -timeout 24h

runsim: $(BINDIR)/runsim
$(BINDIR)/runsim:
	go get github.com/cosmos/tools/cmd/runsim/
	go mod tidy

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test_sim_app_benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

test_sim_app_profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

test_cover:
	@export VERSION=$(VERSION); bash -x tests/test_cover.sh

test_cover_circle:
	@export VERSION="$(git describe --tags --long | sed 's/v\(.*\)/\1/')"

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/cosmos/cosmos-sdk

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)


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
.PHONY: build dist clean test test_unit test_cover lint mocks \
benchmark devdoc_init devdoc devdoc_save devdoc_update runsim \
format test_sim_app_nondeterminism test_sim_modules test_sim_app_fast \
test_sim_app_custom_genesis_fast test_sim_app_custom_genesis_multi_seed \
test_sim_app_multi_seed_short test_sim_app_multi_seed test_sim_app_import_export \
test_sim_benchmark_invariants go-mod-cache
