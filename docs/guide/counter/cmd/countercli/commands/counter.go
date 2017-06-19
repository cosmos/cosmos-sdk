package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	txcmd "github.com/tendermint/light-client/commands/txs"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/docs/guide/counter/plugins/counter"
	btypes "github.com/tendermint/basecoin/types"
)

//CounterTxCmd is the CLI command to execute the counter
//  through the appTx Command
var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "add a vote to the counter",
	Long: `Add a vote to the counter.

You must pass --valid for it to count and the countfee will be added to the counter.`,
	RunE: counterTxCmd,
}

const (
	flagCountFee = "countfee"
	flagValid    = "valid"
)

func init() {
	fs := CounterTxCmd.Flags()
	bcmd.AddAppTxFlags(fs)
	fs.String(flagCountFee, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.Bool(flagValid, false, "Is count valid?")
}

func counterTxCmd(cmd *cobra.Command, args []string) error {
	// Note: we don't support loading apptx from json currently, so skip that

	// Read the app-specific flags
	name, data, err := getAppData()
	if err != nil {
		return err
	}

	// Read the standard app-tx flags
	gas, fee, txInput, err := bcmd.ReadAppTxFlags()
	if err != nil {
		return err
	}

	// Create AppTx and broadcast
	tx := &btypes.AppTx{
		Gas:   gas,
		Fee:   fee,
		Name:  name,
		Input: txInput,
		Data:  data,
	}
	res, err := bcmd.BroadcastAppTx(tx)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(res)
}

func getAppData() (name string, data []byte, err error) {
	countFee, err := btypes.ParseCoins(viper.GetString(flagCountFee))
	if err != nil {
		return
	}
	ctx := counter.CounterTx{
		Valid: viper.GetBool(flagValid),
		Fee:   countFee,
	}

	name = counter.New().Name()
	data = wire.BinaryBytes(ctx)
	return
}
