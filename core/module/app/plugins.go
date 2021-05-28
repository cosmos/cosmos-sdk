package app

import "google.golang.org/grpc"

type PluginProvider interface {
	RegisterPlugins(registrar PluginRegistrar)
}

type PluginRegistrar interface {
	RegisterPluginService(paramType interface{}, desc *grpc.ServiceDesc, impl interface{})
}

func PluginClient(pluginParam interface{}) grpc.ClientConnInterface {
	panic("TODO")
}
