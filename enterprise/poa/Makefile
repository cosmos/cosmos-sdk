###############################################################################
###                                Protobuf                                 ###
###############################################################################

BUF_VERSION=1.59
BUILDER_VERSION=0.17.1

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

build:
	go build -C simapp -o ../build/simd ./simd/

install:
	go install -C simapp ./simd/

build-docker:
	GOOS=linux go build -C simapp -o ../build/simd-linux ./simd/
	docker build -t poa-node .

localnet-clean:
	./scripts/localnet/clean.sh

localnet-genesis:
	./scripts/localnet/genesis.sh

localnet-distribute:
	./scripts/localnet/distribute.sh

localnet-up:
	docker compose up -d

localnet-down:
	docker compose down -v

localnet:
	./scripts/localnet/genesis.sh
	docker compose up -d
	docker compose logs node1 -f

localnet-shell:
	docker compose exec -it node1 bash

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

test-system: build
	mkdir -p ./tests/systemtests/binaries/
	cp ./build/simd ./tests/systemtests/binaries/
	$(MAKE) -C tests/systemtests test

###############################################################################
###                                 Linting                                 ###
###############################################################################
golangci_lint_cmd=golangci-lint
golangci_version=v2.8.0

#? lint: Run latest golangci-lint linter
lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m
.PHONY: lint

#? lint: Run latest golangci-lint linter and apply fixes
lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=15m --fix
.PHONY: lint-fix
