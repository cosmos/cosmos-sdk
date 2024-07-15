module cosmossdk.io/log

go 1.20

require (
	cosmossdk.io/core v0.12.0
	cosmossdk.io/core/testing v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	gotest.tools/v3 v3.5.1
)

require (
	github.com/cosmos/gogoproto v1.5.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/sys v0.22.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace cosmossdk.io/core => ../core

replace cosmossdk.io/core/testing => ../core/testing
