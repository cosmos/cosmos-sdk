package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/types"
)

//commands
var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "Create, sign, and broadcast a transaction to the counter plugin",
	RunE:  counterTxCmd,
}

const (
	flagValid    = "valid"
	flagCountFee = "countfee"
)

func init() {

	CounterTxCmd.Flags().Bool(flagValid, false, "Set valid field in CounterTx")
	CounterTxCmd.Flags().String(flagCountFee, "", "Coins for the counter fee of the format <amt><coin>")

	commands.RegisterTxSubcommand(CounterTxCmd)
	commands.RegisterStartPlugin("counter", func() types.Plugin { return counter.New() })
}

func counterTxCmd(cmd *cobra.Command, args []string) error {

	countFee, err := types.ParseCoins(viper.GetString(flagCountFee))
	if err != nil {
		return err
	}

	counterTx := counter.CounterTx{
		Valid: viper.GetBool(flagValid),
		Fee:   countFee,
	}

	out, err := json.Marshal(counterTx)
	if err != nil {
		return err
	}
	fmt.Println("CounterTx:", string(out))

	data := wire.BinaryBytes(counterTx)
	name := "counter"

	return commands.AppTx(name, data)
}
