package txs

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
)

// nolint
const (
	FlagName    = "name"
	FlagNoSign  = "no-sign"
	FlagIn      = "in"
	FlagPrepare = "prepare"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "tx",
	Short: "Post tx from json input",
	RunE:  doRawTx,
}

func init() {
	RootCmd.PersistentFlags().String(FlagName, "", "name to sign the tx")
	RootCmd.PersistentFlags().Bool(FlagNoSign, false, "don't add a signature")
	RootCmd.PersistentFlags().String(FlagPrepare, "", "file to store prepared tx")
	RootCmd.Flags().String(FlagIn, "", "file with tx in json format")
}

func doRawTx(cmd *cobra.Command, args []string) error {
	raw, err := readInput(viper.GetString(FlagIn))
	if err != nil {
		return err
	}

	// parse the input
	var tx basecoin.Tx
	err = json.Unmarshal(raw, &tx)
	if err != nil {
		return errors.WithStack(err)
	}

	// sign it
	err = SignTx(tx)
	if err != nil {
		return err
	}

	// otherwise, post it and display response
	bres, err := PrepareOrPostTx(tx)
	if err != nil {
		return err
	}
	if bres == nil {
		return nil // successful prep, nothing left to do
	}
	return OutputTx(bres) // print response of the post
}
