package types

import (
	"github.com/gogo/protobuf/grpc"

	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
)

// QueryServer defines the IBC interfaces that the gRPC query server must implement
type QueryServer interface {
	clienttypes.QueryServer
	connectiontypes.QueryServer
	channeltypes.QueryServer
}

// RegisterQueryService registers each individual IBC submodule query service
func RegisterQueryService(server grpc.Server, queryService QueryServer) {
	client.RegisterQueryService(server, queryService)
	connection.RegisterQueryService(server, queryService)
	channel.RegisterQueryService(server, queryService)
}
