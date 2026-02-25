###############################################################################
###                         Group Module Makefile                           ###
###############################################################################

include ../make/enterprise.mk

###############################################################################
###                         Module-Specific Build                           ###
###############################################################################

.PHONY: build build-docker install test-fuzz

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
###                         Module-Specific Tests                           ###
###############################################################################

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
