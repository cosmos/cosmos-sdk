package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryWithdrawAddr(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-addr [account-addr]",
		Short: "Query a distribution withdraw address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := distr.GetDelegatorWithdrawAddrKey(addr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			} else if len(res) == 0 {
				return fmt.Errorf("No validator found with address %s", args[0])
			}

			validator := types.MustUnmarshalValidator(cdc, addr, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				human, err := validator.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(human)

			case "json":
				// parse out the validator
				output, err := codec.MarshalJSONIndent(cdc, validator)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
			}

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}
