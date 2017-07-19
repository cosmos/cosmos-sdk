package txs

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
)

// nolint
const (
	FlagName    = "name"
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
	// TODO: prepare needs to override the SignAndPost somehow to SignAndSave
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

	// Sign if needed and post.  This it the work-horse
	bres, err := SignAndPostTx(tx.Unwrap())
	if err != nil {
		return err
	}
	if err = ValidateResult(bres); err != nil {
		return err
	}

	// Output result
	return OutputTx(bres)
}

func readInput(file string) ([]byte, error) {
	var reader io.Reader
	// get the input stream
	if file == "" || file == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer f.Close()
		reader = f
	}

	// and read it all!
	data, err := ioutil.ReadAll(reader)
	return data, errors.WithStack(err)
}
