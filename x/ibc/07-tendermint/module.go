package tendermint

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/client/cli"
	"github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/client/rest"
	"github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC client
func RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

// GetTxCmd returns the root tx command for the IBC client
func GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}
