package evidence

import (
	"fmt"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	evidencev1beta1 "cosmossdk.io/api/cosmos/evidence/v1beta1"
	"cosmossdk.io/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: evidencev1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Evidence",
					Short:     "Query for evidence by hash or for all (paginated) submitted evidence",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query for specific submitted evidence by hash or query for all (paginated) evidence:

Example:
$ %s query %s DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ %s query %s --page=2 --limit=50
`,
							version.AppName, types.ModuleName, version.AppName, types.ModuleName,
						),
					),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: evidencev1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Evidence",
					Short:     "Evidence transaction subcommands",
				},
			},
		},
	}
}
