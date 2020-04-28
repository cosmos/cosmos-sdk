package module

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
)

type ClientModule interface {
	NewTxCmd(ctx context.CLIContext) *cobra.Command
	NewQueryCmd(ctx context.CLIContext) *cobra.Command
	NewRESTRoutes(ctx context.CLIContext, rtr *mux.Router)
}
