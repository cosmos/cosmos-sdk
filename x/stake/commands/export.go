package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// get the command to export state
func GetCmdExport(load func() (stake.Keeper, sdk.Context), cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export stake module state as JSON to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			keeper, ctx := load()
			genesis := keeper.WriteGenesis(ctx)
			output, err := wire.MarshalJSONIndent(cdc, genesis)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	return cmd
}
