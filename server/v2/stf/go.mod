module cosmossdk.io/server/v2/stf

go 1.23

replace cosmossdk.io/core => ../../../core

require (
	cosmossdk.io/core v0.11.0
	github.com/cosmos/gogoproto v1.7.0
	github.com/tidwall/btree v1.7.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/cosmos/gogoproto => github.com/cosmos/gogoproto v1.6.1-0.20240809124342-d6a57064ada0
