module cosmossdk.io/core

go 1.20

require github.com/cosmos/gogoproto v1.6.0

require (
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

// Version tagged too early and incompatible with v0.50 (latest at the time of tagging)
retract v0.12.0
