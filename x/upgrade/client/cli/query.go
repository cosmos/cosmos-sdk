package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/spf13/cobra"
)

// GetQueryCmd creates a query sub-command for the upgrade module using cmdName as the name of the sub-command.
func GetQueryCmd(cmdName string, storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   cmdName,
		Short: "get upgrade plan (if one exists)",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore([]byte(upgrade.PlanKey), storeName)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("no upgrade scheduled")
			}

			var plan upgrade.Plan
			err = cdc.UnmarshalBinaryBare(res, &plan)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(plan)
		},
	}
}
