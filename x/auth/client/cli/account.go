package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAccountCmd returns a query account that will display the state of the
// account at a given address.
// nolint: unparam
func GetAccountCmd(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// find the key to look up the account
			addr := args[0]

			key, err := sdk.AccAddressFromBech32(addr)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			if err = cliCtx.EnsureAccountExistsFromAddr(key); err != nil {
				return err
			}

			acc, err := cliCtx.GetAccount(key)
			if err != nil {
				return err
			}

			var output []byte
			if cliCtx.Indent {
				output, err = cdc.MarshalJSONIndent(acc, "", "  ")
			} else {
				output, err = cdc.MarshalJSON(acc)
			}
			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}

	// Add the flags here and return the command
	return client.GetCommands(cmd)[0]
}
