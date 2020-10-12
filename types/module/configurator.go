package module

import "github.com/gogo/protobuf/grpc"

type Configurator interface {
	QueryServer() grpc.Server
}

type configurator struct {
	queryServer grpc.Server
}

func NewConfigurator(queryServer grpc.Server) Configurator {
	return configurator{queryServer: queryServer}
}

var _ Configurator = configurator{}

func (c configurator) QueryServer() grpc.Server {
	return c.queryServer
}
