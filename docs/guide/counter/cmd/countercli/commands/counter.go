package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	txcmd "github.com/tendermint/basecoin/client/commands/txs"

	"github.com/tendermint/basecoin"
	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/docs/guide/counter/plugins/counter"
	"github.com/tendermint/basecoin/modules/coin"
)

//CounterTxCmd is the CLI command to execute the counter
//  through the appTx Command
var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "add a vote to the counter",
	Long: `Add a vote to the counter.

You must pass --valid for it to count and the countfee will be added to the counter.`,
	RunE: counterTx,
}

// nolint - flags names
const (
	FlagCountFee = "countfee"
	FlagValid    = "valid"
)

func init() {
	fs := CounterTxCmd.Flags()
	fs.String(FlagCountFee, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.Bool(FlagValid, false, "Is count valid?")
}

// TODO: counterTx is very similar to the sendtx one,
// maybe we can pull out some common patterns?
func counterTx(cmd *cobra.Command, args []string) error {
	tx, err := readCounterTxFlags()
	if err != nil {
		return err
	}

	tx, err = bcmd.Middleware.Wrap(tx)
	if err != nil {
		return err
	}

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(tx.Unwrap())
	if err != nil {
		return err
	}
	if err = bcmd.ValidateResult(bres); err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readCounterTxFlags() (tx basecoin.Tx, err error) {
	feeCoins, err := coin.ParseCoins(viper.GetString(FlagCountFee))
	if err != nil {
		return tx, err
	}

	tx = counter.NewTx(viper.GetBool(FlagValid), feeCoins)
	return tx, nil
}
