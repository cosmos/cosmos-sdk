LINKER_FLAGS:="-X github.com/cosmos/cosmos-sdk/client/commands.CommitHash=`git rev-parse --short HEAD`"

install:
	@go install -ldflags $(LINKER_FLAGS) ./cmd/...

test: test_unit test_cli

test_unit:
	@go test `glide novendor`

test_cli:
	./tests/cli/keys.sh
	./tests/cli/rpc.sh
	./tests/cli/init.sh
	./tests/cli/basictx.sh
	./tests/cli/roles.sh
	./tests/cli/restart.sh
	./tests/cli/rest.sh
	./tests/cli/ibc.sh

.PHONY: install test test_unit test_cli
