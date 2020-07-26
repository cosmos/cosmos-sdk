package types

import (
	"github.com/gogo/protobuf/grpc"

	connection "github.com/KiraCore/cosmos-sdk/x/ibc/03-connection"
	connectiontypes "github.com/KiraCore/cosmos-sdk/x/ibc/03-connection/types"
	channel "github.com/KiraCore/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/KiraCore/cosmos-sdk/x/ibc/04-channel/types"
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
