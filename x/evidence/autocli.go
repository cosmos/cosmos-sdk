package evidence

import (
	"fmt"

	evidencev1beta1 "cosmossdk.io/api/cosmos/evidence/v1beta1"
	autocli "cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: evidencev1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod:      "Evidence",
					Use:            "evidence [hash]",
					Short:          "Query for evidence by hash",
					Example:        fmt.Sprintf("%s query evidence evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660", version.AppName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "hash"}},
				},
				{
					RpcMethod: "AllEvidence",
					Use:       "list",
					Short:     "Query all (paginated) submitted evidence",
					Example:   fmt.Sprintf("%s query evidence list --page=2 --page-limit=50", version.AppName),
				},
			},
		},
	}
}
