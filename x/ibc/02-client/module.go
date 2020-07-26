package client

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/client/cli"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/client/rest"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC client
func RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

// GetQueryCmd returns no root query command for the IBC client
func GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}
