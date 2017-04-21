package main

import (
	"fmt"

	"github.com/spf13/cobra"
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

//flags
var (
	validFlag    bool
	countFeeFlag string
)

func init() {

	CounterTxCmd.Flags().BoolVar(&validFlag, "valid", false, "Set valid field in CounterTx")
	CounterTxCmd.Flags().StringVar(&countFeeFlag, "countfee", "", "Coins for the counter fee of the format <amt><coin>")

	commands.RegisterTxSubcommand(CounterTxCmd)
	commands.RegisterStartPlugin("counter", func() types.Plugin { return counter.New() })
}

func counterTxCmd(cmd *cobra.Command, args []string) error {

	countFee, err := types.ParseCoins(countFeeFlag)
	if err != nil {
		return err
	}

	counterTx := counter.CounterTx{
		Valid: validFlag,
		Fee:   countFee,
	}

	fmt.Println("CounterTx:", string(wire.JSONBytes(counterTx)))

	data := wire.BinaryBytes(counterTx)
	name := "counter"

	return commands.AppTx(name, data)
}
