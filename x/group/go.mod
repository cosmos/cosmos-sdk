go 1.15

module github.com/cosmos/cosmos-sdk/x/group

require (
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/gogo/protobuf v1.3.3
	github.com/regen-network/cosmos-proto v0.3.1
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/cosmos/cosmos-sdk => ../../
