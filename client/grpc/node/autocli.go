package node

import (
	nodev1beta1 "cosmossdk.io/api/cosmos/base/node/v1beta1"
	autocli "cosmossdk.io/core/autocli"
)

var ServiceAutoCLIDescriptor = &autocli.ServiceCommandDescriptor{
	Service: nodev1beta1.Service_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocli.RpcCommandOptions{
		{
			RpcMethod: "Config",
			Use:       "config",
			Short:     "Query the current node config",
		},
		{
			RpcMethod: "Status",
			Use:       "status",
			Short:     "Query the current node status",
		},
	},
}

// NewNodeCommands is a fake `appmodule.Module` to be considered as a module
// and be added in AutoCLI.
func NewNodeCommands() *nodeModule {
	return &nodeModule{}
}

type nodeModule struct{}

func (m nodeModule) IsOnePerModuleType() {}
func (m nodeModule) IsAppModule()        {}

func (m nodeModule) Name() string {
	return "node"
}

func (m nodeModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: ServiceAutoCLIDescriptor,
	}
}
