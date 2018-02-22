package main

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
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
)

func getAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account <address>",
		Short: "Query account balance",
		RunE:  getAccount,
	}
	return cmd
}

func getAccount(cmd *cobra.Command, args []string) error {
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

	// TODO: make the store name a variable in getAccountCmd?
	path := "/main/key"

	uri := viper.GetString(flagNode)
	if uri == "" {
		return errors.New("Must define which node to query with --node")
	}
	node := client.GetNode(uri)

	opts := rpcclient.ABCIQueryOptions{
		Height: viper.GetInt64(flagHeight),
		// Trusted: viper.GetBool(flagTrustNode),
		Trusted: true,
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

	// TODO: better
	// fmt.Printf("Account: %#v\n", acct)
	output, err := json.MarshalIndent(acct, "", "  ")
	// output, err := json.MarshalIndent(acct.BaseAccount.Coins, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
