.PHONY: all build install clean test

BINARY=tajeord
GOBIN=$(shell go env GOPATH)/bin

all: build

build:
	@echo "Building ${BINARY}..."
	@go build -o build/${BINARY} ./cmd/tajeord

install:
	@echo "Installing ${BINARY}..."
	@go install ./cmd/tajeord

clean:
	@echo "Cleaning..."
	@rm -rf build/

test:
	@echo "Running tests..."
	@go test ./...

lint:
	@echo "Running linter..."
	@golangci-lint run

.PHONY: proto
proto:
	@echo "Generating protobuf files..."
	@buf generate

.PHONY: format
format:
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy 