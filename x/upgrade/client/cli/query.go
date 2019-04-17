package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/spf13/cobra"
)

// GetPlanCmd returns the query upgrade plan command
func GetPlanCmd(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
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

// GetHeightCmd returns the height at which a completed upgrade was applied
func GetHeightCmd(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "height [upgrade-name]",
		Short: "get the height at which a completed upgrade was applied",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			name := args[0]

			res, err := cliCtx.QueryStore(upgrade.DoneHeightKey(name), storeName)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("no upgrade found")
			}

			var height int64
			err = cdc.UnmarshalBinaryBare(res, &height)
			if err != nil {
				return err
			}
			fmt.Println(height)
			return nil
		},
	}
}
