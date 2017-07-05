package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"

	"github.com/tendermint/basecoin/docs/guide/counter/plugins/counter"
	"github.com/tendermint/basecoin/txs"
	btypes "github.com/tendermint/basecoin/types"
)

//CounterTxCmd is the CLI command to execute the counter
//  through the appTx Command
var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "add a vote to the counter",
	Long: `Add a vote to the counter.

You must pass --valid for it to count and the countfee will be added to the counter.`,
	RunE: doCounterTx,
}

// nolint - flags names
const (
	FlagCountFee = "countfee"
	FlagValid    = "valid"
	FlagSequence = "sequence" // FIXME: currently not supported...
)

func init() {
	fs := CounterTxCmd.Flags()
	fs.String(FlagCountFee, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.Bool(FlagValid, false, "Is count valid?")
	fs.Int(FlagSequence, -1, "Sequence number for this transaction")
}

// TODO: doCounterTx is very similar to the sendtx one,
// maybe we can pull out some common patterns?
func doCounterTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	var tx basecoin.Tx
	found, err := txcmd.LoadJSON(&tx)
	if err != nil {
		return err
	}
	if !found {
		tx, err = readCounterTxFlags()
	}
	if err != nil {
		return err
	}

	// TODO: make this more flexible for middleware
	// add the chain info
	tx = txs.NewChain(commands.GetChainID(), tx)
	stx := txs.NewSig(tx)

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(stx)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readCounterTxFlags() (tx basecoin.Tx, err error) {
	feeCoins, err := btypes.ParseCoins(viper.GetString(FlagCountFee))
	if err != nil {
		return tx, err
	}

	tx = counter.NewTx(viper.GetBool(FlagValid), feeCoins, viper.GetInt(FlagSequence))
	return tx, nil
}
