package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"   // XXX: not good
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types" // XXX: not good
)

// GetAccountCmd returns a query account that will display the
// state of the account at a given address
func GetAccountCmd(storeName string) *cobra.Command {
	cmd := acctCmd{storeName}

	return &cobra.Command{
		Use:   "account <address>",
		Short: "Query account balance",
		RunE:  cmd.get,
	}
}

type acctCmd struct {
	storeName string
}

func (a acctCmd) get(cmd *cobra.Command, args []string) error {
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
	path := fmt.Sprintf("/%s/key", a.storeName)

	uri := viper.GetString(client.FlagNode)
	if uri == "" {
		return errors.New("Must define which node to query with --node")
	}
	node := client.GetNode(uri)

	opts := rpcclient.ABCIQueryOptions{
		Height:  viper.GetInt64(client.FlagHeight),
		Trusted: viper.GetBool(client.FlagTrustNode),
	}
	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return err
	}
	resp := result.Response
	if resp.Code != uint32(0) {
		return errors.Errorf("Query failed: (%d) %s", resp.Code, resp.Log)
	}

	// parse out the value
	acct := new(types.AppAccount)
	cdc := app.MakeTxCodec()
	err = cdc.UnmarshalBinary(resp.Value, acct)
	if err != nil {
		return err
	}

	// print out whole account or just coins?
	output, err := json.MarshalIndent(acct, "", "  ")
	// output, err := json.MarshalIndent(acct.BaseAccount.Coins, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
