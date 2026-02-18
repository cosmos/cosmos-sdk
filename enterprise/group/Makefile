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

BUILD_DIR ?= ./build

# Build group module and simd binary
build:
	@echo "Building group module..."
	@go build ./...
	@echo "Building simd..."
	@mkdir -p $(BUILD_DIR)
	@go build -C simapp -o ../$(BUILD_DIR)/simd ./simd

# Install simd binary to GOPATH/bin
install: build
	@cp $(BUILD_DIR)/simd $(shell go env GOPATH)/bin/simd

build-docker: build
	@echo "Building Linux binary for Docker..."
	@GOOS=linux go build -C simapp -o ../$(BUILD_DIR)/simd-linux ./simd
	@echo "Building Docker image..."
	@docker build -t group-simd .

###############################################################################
###                                  Tests                                  ###
###############################################################################

.PHONY: test test-verbose test-race test-cover test-fuzz

#? test: Run all tests
test:
	@echo "--> Running tests"
	@go test ./...

#? test-fuzz: Run types-level fuzz tests (30s each)
test-fuzz:
	@echo "--> Running fuzz tests (group internal math)"
	@go test -fuzz=FuzzNewDecFromString -fuzztime=30s ./x/group/internal/math
	@go test -fuzz=FuzzNewPositiveDecFromString -fuzztime=30s ./x/group/internal/math
	@go test -fuzz=FuzzNewNonNegativeDecFromString -fuzztime=30s ./x/group/internal/math
	@echo "--> Running fuzz tests (group types)"
	@go test -fuzz=FuzzTallyResultGetCounts -fuzztime=30s ./x/group
	@go test -fuzz=FuzzTallyResultTotalCounts -fuzztime=30s ./x/group
	@go test -fuzz=FuzzVoteOptionFromString -fuzztime=30s ./x/group
	@echo "--> Fuzz tests completed"

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
