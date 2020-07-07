package types

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/gogo/protobuf/grpc"
)

// QueryServer defines the IBC interfaces that the gRPC query server must implement
type QueryServer interface {
	connectiontypes.QueryServer
	channeltypes.QueryServer
}

// RegisterQueryService registers each individual IBC submodule query service
func RegisterQueryService(server grpc.Server, queryService QueryServer) {
	connection.RegisterQueryService(server, queryService)
	channel.RegisterQueryService(server, queryService)
}
