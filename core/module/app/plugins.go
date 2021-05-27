package app

import "google.golang.org/grpc"

type HasPlugins interface {
	Module

	RegisterPlugins(registrar grpc.ServiceRegistrar)
}

type PluginRegistrar interface {
	RegisterPluginService(desc *grpc.ServiceDesc, impl interface{})
}
