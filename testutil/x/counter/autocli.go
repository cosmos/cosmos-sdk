package counter

import (
	counterv1 "cosmossdk.io/api/cosmos/counter/v1"
	autocli "cosmossdk.io/core/autocli"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: counterv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "GetCount",
					Use:       "count",
					Short:     "Query the current counter value",
				},
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service: counterv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod:      "IncreaseCount",
					Use:            "increase-count <count>",
					Alias:          []string{"increase-counter", "increase", "inc", "bump"},
					Short:          "Increase the counter by the specified amount",
					PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "count"}},
				},
			},
		},
	}
}
