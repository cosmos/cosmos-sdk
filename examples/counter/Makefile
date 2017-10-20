COMMIT_HASH := $(shell git rev-parse --short HEAD)
LINKER_FLAGS:="-X github.com/cosmos/cosmos-sdk/client/commands.CommitHash=${COMMIT_HASH}"

install:
	@go install -ldflags $(LINKER_FLAGS) ./cmd/...

test: test_unit test_cli

test_unit:
	@go test `glide novendor`

test_cli:
	./tests/cli/counter.sh

.PHONY: install test test_unit test_cli