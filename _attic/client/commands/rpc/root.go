package rpc

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/go-wire/data"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands"
)

const (
	FlagDelta  = "delta"
	FlagHeight = "height"
	FlagMax    = "max"
	FlagMin    = "min"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "rpc",
	Short: "Query the tendermint rpc, validating everything with a proof",
}

// TODO: add support for subscribing to events????
func init() {
	RootCmd.AddCommand(
		statusCmd,
		infoCmd,
		genesisCmd,
		validatorsCmd,
		blockCmd,
		commitCmd,
		headersCmd,
		waitCmd,
	)
}

func getSecureNode() (rpcclient.Client, error) {
	// First, connect a client
	c := commands.GetNode()
	cert, err := commands.GetCertifier()
	if err != nil {
		return nil, err
	}
	return client.SecureClient(c, cert), nil
}

// printResult just writes the struct to the console, returns an error if it can't
func printResult(res interface{}) error {
	// TODO: handle text mode
	// switch viper.Get(cli.OutputFlag) {
	// case "text":
	// case "json":
	json, err := data.ToJSON(res)
	if err != nil {
		return err
	}
	fmt.Println(string(json))
	return nil
}
