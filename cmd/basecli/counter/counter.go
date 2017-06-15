package counter

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	txcmd "github.com/tendermint/light-client/commands/txs"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/plugins/counter"
	btypes "github.com/tendermint/basecoin/types"
)

var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "add a vote to the counter",
	Long: `Add a vote to the counter.

You must pass --valid for it to count and the countfee will be added to the counter.`,
	RunE: doCounterTx,
}

const (
	CountFeeFlag = "countfee"
	ValidFlag    = "valid"
)

func init() {
	fs := CounterTxCmd.Flags()
	bcmd.AddAppTxFlags(fs)
	fs.String(CountFeeFlag, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.Bool(ValidFlag, false, "Is count valid?")
}

func doCounterTx(cmd *cobra.Command, args []string) error {
	tx := new(btypes.AppTx)
	// Note: we don't support loading apptx from json currently, so skip that

	// read the standard flags
	err := bcmd.ReadAppTxFlags(tx)
	if err != nil {
		return err
	}

	// now read the app-specific flags
	err = readCounterFlags(tx)
	if err != nil {
		return err
	}

	app := bcmd.WrapAppTx(tx)
	app.AddSigner(txcmd.GetSigner())

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(app)
	if err != nil {
		return err
	}

	// output result
	return txcmd.OutputTx(bres)
}

// readCounterFlags sets the app-specific data in the AppTx
func readCounterFlags(tx *btypes.AppTx) error {
	countFee, err := btypes.ParseCoins(viper.GetString(CountFeeFlag))
	if err != nil {
		return err
	}
	ctx := counter.CounterTx{
		Valid: viper.GetBool(ValidFlag),
		Fee:   countFee,
	}

	tx.Name = counter.New().Name()
	tx.Data = wire.BinaryBytes(ctx)
	return nil
}
