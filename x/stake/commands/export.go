package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/wire" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// get the command to export state
func GetCmdExport(export func() (stake.GenesisState, error), cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-stake",
		Short: "Export stake module state as JSON to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			genesis, err := export()
			if err != nil {
				return err
			}
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
