package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/docs/guide/counter/plugins/counter"
	"github.com/cosmos/cosmos-sdk/modules/coin"
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

func counterTx(cmd *cobra.Command, args []string) error {
	tx, err := readCounterTxFlags()
	if err != nil {
		return err
	}
	return txcmd.DoTx(tx)
}

func readCounterTxFlags() (tx sdk.Tx, err error) {
	feeCoins, err := coin.ParseCoins(viper.GetString(FlagCountFee))
	if err != nil {
		return tx, err
	}

	tx = counter.NewTx(viper.GetBool(FlagValid), feeCoins)
	return tx, nil
}
