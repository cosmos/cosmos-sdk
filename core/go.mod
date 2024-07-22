module cosmossdk.io/core

go 1.20

require (
	github.com/cosmos/gogoproto v1.5.0
	google.golang.org/grpc v1.64.1
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

// Version tagged too early and incompatible with v0.50 (latest at the time of tagging)
retract v0.12.0
