package solomachine

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

// Name returns the solo machine client name.
func Name() string {
	return types.SubModuleName
}

// RegisterRESTRoutes returns nil. IBC does not support legacy REST routes.
func RegisterRESTRoutes(client.Context, *mux.Router, string) {}

// GetTxCmd returns the root tx command for the solo machine client.
func GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}
