package fee_delegation

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-delegation",
		Short: "Querying commands for the delegation module",
	}

	cmd.AddCommand(client.GetCommands(
		GetCmdGetFeeAllowances(queryRoute, cdc),
	)...)

	return cmd
}

func GetCmdGetFeeAllowances(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "allowances [address]",
		Short: "get fee allowances granted to this address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := args[0]

			route := fmt.Sprintf("custom/delegation/%s/%s", QueryGetFeeAllowances, id)
			res, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}
