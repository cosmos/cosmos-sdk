package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// GetAccountCmd for the auth.BaseAccount type
func GetAccountCmdDefault(storeName string, cdc *wire.Codec) *cobra.Command {
	return GetAccountCmd(storeName, cdc, GetAccountDecoder(cdc))
}

// Get account decoder for auth.DefaultAccount
func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (acct auth.Account, err error) {
		// acct := new(auth.BaseAccount)
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}
		return acct, err
	}
}

// GetAccountCmd returns a query account that will display the
// state of the account at a given address
func GetAccountCmd(storeName string, cdc *wire.Codec, decoder auth.AccountDecoder) *cobra.Command {
	return &cobra.Command{
		Use:   "account [address]",
		Short: "Query account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// find the key to look up the account
			addr := args[0]

			key, err := sdk.GetAccAddressBech32(addr)
			if err != nil {
				return err
			}

			// perform query
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore(auth.AddressStoreKey(key), storeName)
			if err != nil {
				return err
			}

			// Check if account was found
			if res == nil {
				return sdk.ErrUnknownAddress("No account with address " + addr +
					" was found in the state.\nAre you sure there has been a transaction involving it?")
			}

			// decode the value
			account, err := decoder(res)
			if err != nil {
				return err
			}

			// print out whole account
			output, err := wire.MarshalJSONIndent(cdc, account)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
}
