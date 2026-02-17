###############################################################################
###                                Protobuf                                 ###
###############################################################################

BUF_VERSION=1.64
BUILDER_VERSION=0.18.0

.PHONY: proto-all proto-format proto-gen proto-lint

proto-all: proto-format proto-lint proto-gen

proto-format:
	@echo "🤖 Running protobuf formatter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write
	@echo "✅ Completed protobuf formatting!"

proto-gen:
	@echo "🤖 Generating code from protobuf..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh ./scripts/protogen.sh
	@echo "✅ Completed code generation!"

proto-lint:
	@echo "🤖 Running protobuf linter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint
	@echo "✅ Completed protobuf linting!"

license:
	@echo "🤖 Adding license headers to all source files..."
	@./scripts/add-license.sh
	@echo "✅ License headers added!"

.PHONY: build build-docker install license

BUILD_DIR ?= $(shell go env GOPATH)/bin

# Build group module and simd binary
build:
	@echo "Building group module..."
	@go build ./...
	@echo "Building simd..."
	@go build -C simapp -o $(BUILD_DIR)/simd ./simd

# Install simd binary to GOPATH/bin
install: build

build-docker:
	@echo "Docker build not yet implemented"

###############################################################################
###                                  Tests                                  ###
###############################################################################

.PHONY: test test-verbose test-race test-cover

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

# test-system requires simapp - not yet implemented
test-system:
	@echo "System tests require simapp - not yet implemented"

###############################################################################
###                                 Linting                                 ###
###############################################################################
golangci_lint_cmd=golangci-lint
golangci_version=v2.9.0

#? lint: Run latest golangci-lint linter
lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m
.PHONY: lint

#? lint-fix: Run latest golangci-lint linter and apply fixes
lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m --fix
.PHONY: lint-fix
