package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types" // XXX: not good

	"github.com/cosmos/cosmos-sdk/x/bank"
)

// GetAccountCmd returns a query account that will display the
// state of the account at a given address
func GetAccountCmd(storeName string) *cobra.Command {
	return &cobra.Command{
		Use:   "account <address>",
		Short: "Query account balance",
		RunE:  newRunner(storeName).cmd,
	}
}

type runner struct {
	storeName string
}

func newRunner(storeName string) runner {
	return runner{
		storeName: storeName,
	}
}

func (r runner) cmd(cmd *cobra.Command, args []string) error {
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

	res, err := client.Query(key, r.storeName)

	// parse out the value
	acct := new(types.AppAccount)
	cdc := wire.NewCodec()
	bank.RegisterWire(cdc)

	err = cdc.UnmarshalBinary(res, acct)
	if err != nil {
		return err
	}

	// print out whole account
	output, err := json.MarshalIndent(acct, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}
