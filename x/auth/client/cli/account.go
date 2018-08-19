package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// GetAccountCmdDefault invokes the GetAccountCmd for the auth.BaseAccount type.
func GetAccountCmdDefault(storeName string, cdc *wire.Codec) *cobra.Command {
	return GetAccountCmd(storeName, cdc, GetAccountDecoder(cdc))
}

// GetAccountDecoder gets the account decoder for auth.DefaultAccount.
func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (acct auth.Account, err error) {
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}

		return acct, err
	}
}

// GetAccountCmd returns a query account that will display the state of the
// account at a given address.
func GetAccountCmd(storeName string, cdc *wire.Codec, decoder auth.AccountDecoder) *cobra.Command {
	return &cobra.Command{
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
				WithAccountDecoder(decoder)

			if err := cliCtx.EnsureAccountExistsFromAddr(key); err != nil {
				return err
			}

			acc, err := cliCtx.GetAccount(key)
			if err != nil {
				return err
			}

			output, err := wire.MarshalJSONIndent(cdc, acc)
			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}
}
