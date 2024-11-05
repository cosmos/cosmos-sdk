module github.com/cosmos/cosmos-sdk/x/benchmark

go 1.23.2

require (
	cosmossdk.io/api v0.7.0
	cosmossdk.io/core v1.0.0-alpha.5
	cosmossdk.io/depinject v1.0.0
	github.com/cosmos/gogoproto v1.7.0
)

require (
	cosmossdk.io/schema v0.3.0 // indirect
	github.com/cosmos/cosmos-proto v1.0.0-beta.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240930140551-af27646dc61f // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace cosmossdk.io/api => ../../api
