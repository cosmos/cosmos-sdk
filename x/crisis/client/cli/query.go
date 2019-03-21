package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/crisis"
)

// GetCmdQuerySigningInfo implements the command to query signing info.
func GetCmdQuerySigningInfo(storeName string, cdc *codec.Codec, routes crisis.Routes) *cobra.Command {
	return &cobra.Command{
		Use:   "invariant-routes",
		Short: "list the available invariant routes available",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			return cliCtx.PrintOutput(routes)
		},
	}
}
