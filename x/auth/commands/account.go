package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/client" // XXX: not good
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// GetAccountCmd for the auth.BaseAccount type
func GetAccountCmdDefault(storeName string, cdc *wire.Codec) *cobra.Command {
	return GetAccountCmd(storeName, cdc, getParseAccount(cdc))
}

func getParseAccount(cdc *wire.Codec) sdk.ParseAccount {
	return func(accBytes []byte) (sdk.Account, error) {
		acct := new(auth.BaseAccount)
		err := cdc.UnmarshalBinary(accBytes, acct)
		return acct, err
	}
}

// GetAccountCmd returns a query account that will display the
// state of the account at a given address
func GetAccountCmd(storeName string, cdc *wire.Codec, parser sdk.ParseAccount) *cobra.Command {
	cmdr := commander{
		storeName,
		cdc,
		parser,
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
	parser    sdk.ParseAccount
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
	key := crypto.Address(bz)

	res, err := client.Query(key, c.storeName)

	// parse out the value
	account, err := c.parser(res)
	if err != nil {
		return err
	}

	// print out whole account
	output, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}
