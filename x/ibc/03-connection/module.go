package connection

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// Name returns the IBC connection ICS name
func Name() string {
	return SubModuleName
}

// GetTxCmd returns the root tx command for the IBC connections.
func GetTxCmd(clientCtx client.Context) *cobra.Command {
	return cli.NewTxCmd(clientCtx)
}

// GetQueryCmd returns no root query command for the IBC connections.
func GetQueryCmd(clientCtx client.Context) *cobra.Command {
	return cli.GetQueryCmd(clientCtx)
}

// RegisterRESTRoutes registers the REST routes for the IBC connections.
func RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

// RegisterQueryService registers the gRPC query service for IBC connections.
func RegisterQueryService(server grpc.Server, k keeper.Keeper) {
	types.RegisterQueryServer(server, k)
}
