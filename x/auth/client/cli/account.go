package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// GetAccountCmd for the auth.BaseAccount type
func GetAccountCmdDefault(storeName string, cdc *wire.Codec) *cobra.Command {
	return GetAccountCmd(storeName, cdc, GetAccountDecoder(cdc))
}

// Get account decoder for auth.DefaultAccount
func GetAccountDecoder(cdc *wire.Codec) sdk.AccountDecoder {
	return func(accBytes []byte) (acct sdk.Account, err error) {
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
func GetAccountCmd(storeName string, cdc *wire.Codec, decoder sdk.AccountDecoder) *cobra.Command {
	cmdr := commander{
		storeName,
		cdc,
		decoder,
	}
	return &cobra.Command{
		Use:   "account <address>",
		Short: "Query account balance",
		RunE:  cmdr.getAccountCmd,
	}
}

type commander struct {
	storeName string
	cdc       *wire.Codec
	decoder   sdk.AccountDecoder
}

func (c commander) getAccountCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide an account name")
	}

	// find the key to look up the account
	addr := args[0]
	bz, err := hex.DecodeString(addr)
	if err != nil {
		return err
	}
	key := sdk.Address(bz)

	ctx := context.NewCoreContextFromViper()

	res, err := ctx.Query(key, c.storeName)
	if err != nil {
		return err
	}

	// decode the value
	account, err := c.decoder(res)
	if err != nil {
		return err
	}

	// print out whole account
	output, err := wire.MarshalJSONIndent(c.cdc, account)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}
