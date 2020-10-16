package module

import "github.com/gogo/protobuf/grpc"

// Configurator provides the hooks to allow modules to configure and register
// their services in the RegisterServices method. It is designed to eventually
// support module object capabilities isolation as described in
// https://github.com/cosmos/cosmos-sdk/issues/7093
type Configurator interface {
	MsgServer() grpc.Server
	QueryServer() grpc.Server
}

type configurator struct {
	msgServer   grpc.Server
	queryServer grpc.Server
}

// NewConfigurator returns a new Configurator instance
func NewConfigurator(msgServer grpc.Server, queryServer grpc.Server) Configurator {
	return configurator{msgServer: msgServer, queryServer: queryServer}
}

var _ Configurator = configurator{}

func (c configurator) MsgServer() grpc.Server {
	return c.msgServer
}

// QueryServer implements the Configurator.QueryServer method
func (c configurator) QueryServer() grpc.Server {
	return c.queryServer
}
