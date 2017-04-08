package main

import (
	"fmt"

	"github.com/spf13/cobra"
	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
)

//commands
var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "Create, sign, and broadcast a transaction to the counter plugin",
	Run:   counterTxCmd,
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

func counterTxCmd(cmd *cobra.Command, args []string) {

	countFee, err := commands.ParseCoins(countFeeFlag)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	counterTx := counter.CounterTx{
		Valid: validFlag,
		Fee:   countFee,
	}

	fmt.Println("CounterTx:", string(wire.JSONBytes(counterTx)))

	data := wire.BinaryBytes(counterTx)
	name := "counter"

	commands.AppTx(name, data)
}
