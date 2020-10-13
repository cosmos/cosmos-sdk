package module

import "github.com/gogo/protobuf/grpc"

// Configurator provides the hooks to allow modules to configure and register
// their services in the RegisterServices method. It is designed to eventually
// support module object capabilities isolation as described in
// https://github.com/cosmos/cosmos-sdk/issues/7093
type Configurator interface {
	QueryServer() grpc.Server
}

type configurator struct {
	queryServer grpc.Server
}

// NewConfigurator returns a new Configurator instance
func NewConfigurator(queryServer grpc.Server) Configurator {
	return configurator{queryServer: queryServer}
}

var _ Configurator = configurator{}

// QueryServer implements the Configurator.QueryServer method
func (c configurator) QueryServer() grpc.Server {
	return c.queryServer
}
