module cosmossdk.io/server/v2/appmanager

go 1.21

replace (
	cosmossdk.io/core => ../../../core
	cosmossdk.io/server/v2/core => ../core
	cosmossdk.io/server/v2/stf => ../stf
	github.com/cosmos/iavl => github.com/cosmos/iavl v1.0.0-beta.1.0.20240125174944-11ba4961dae9
)

require (
	cosmossdk.io/core v0.12.0
	cosmossdk.io/log v1.3.1
	cosmossdk.io/server/v2/core v0.0.0-00010101000000-000000000000
)

require (
	github.com/cosmos/gogoproto v1.4.11 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.32.0 // indirect
	golang.org/x/exp v0.0.0-20240205201215-2c58cdc269a3 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)
