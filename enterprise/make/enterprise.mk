# enterprise/make/enterprise.mk - Shared Makefile for enterprise modules
# Include from module Makefiles: include ../make/enterprise.mk
#
# Optional overrides (set before include):
#   DOCKER_IMAGE: Docker image tag for build-docker (default: $(ENTERPRISE_MODULE)-simd)
#   BUILD_MODULE_FIRST: If set, run 'go build ./...' before building simd (group uses this)
#   BUILD_DIR: Default ./build

###############################################################################
###                              Variables                                  ###
###############################################################################

# Derived from CURDIR when in enterprise/group or enterprise/poa
ENTERPRISE_MODULE ?= $(notdir $(CURDIR))
DOCKER_IMAGE ?= $(ENTERPRISE_MODULE)-simd

###############################################################################
###                              Tool Versions                              ###
###############################################################################

BUF_VERSION=1.66
BUILDER_VERSION=0.18.0
golangci_version=v2.10.1

all: build lint test

###############################################################################
###                                Protobuf                                 ###
###############################################################################

.PHONY: proto-all proto-format proto-gen proto-lint license

proto-all: proto-format proto-lint proto-gen

proto-format:
	@echo "🤖 Running protobuf formatter..."
	@docker run --rm --volume "$(CURDIR)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write
	@echo "✅ Completed protobuf formatting!"

proto-gen:
	@echo "🤖 Generating code from protobuf..."
	@docker run --rm --volume "$(CURDIR)/../..":/repo --volume "$(CURDIR)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh /repo/enterprise/scripts/protogen.sh $(ENTERPRISE_MODULE)
	@echo "✅ Completed code generation!"

proto-lint:
	@echo "🤖 Running protobuf linter..."
	@docker run --rm --volume "$(CURDIR)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint
	@echo "✅ Completed protobuf linting!"

license:
	@echo "🤖 Adding license headers to all source files..."
	@sh $(CURDIR)/../scripts/add-license.sh $(CURDIR) enterprise/$(ENTERPRISE_MODULE)
	@echo "✅ License headers added!"

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_DIR ?= ./build

.PHONY: build install build-docker

build:
ifdef BUILD_MODULE_FIRST
	@echo "Building module..."
	@go build ./...
endif
	@echo "Building simd..."
	@mkdir -p $(BUILD_DIR)
	@go build -C simapp -o ../$(BUILD_DIR)/simd ./simd

install: build
	@cp $(BUILD_DIR)/simd $(shell go env GOPATH)/bin/simd

build-docker: build
	@echo "Building Linux binary for Docker..."
	@GOOS=linux go build -C simapp -o ../$(BUILD_DIR)/simd-linux ./simd
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

###############################################################################
###                                  Tests                                  ###
###############################################################################
golangci_lint_cmd=golangci-lint

.PHONY: test test-verbose test-cover test-system

#? test: Run all tests
test:
	@echo "--> Running tests"
	@go test ./...

#? test-verbose: Run all tests with verbose output
test-verbose:
	@echo "--> Running tests (verbose)"
	@go test -v ./...

#? test-cover: Run all tests with coverage report
test-cover:
	@echo "--> Running tests (with coverage)"
	@go test -cover ./...

test-system: build
	@mkdir -p ./tests/systemtests/binaries
	@cp $(BUILD_DIR)/simd ./tests/systemtests/binaries/
	@$(MAKE) -C tests/systemtests test

###############################################################################
###                                 Linting                                 ###
###############################################################################

.PHONY: lint lint-fix

#? lint: Run golangci-lint linter
lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m

#? lint-fix: Run golangci-lint linter and apply fixes
lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m --fix
