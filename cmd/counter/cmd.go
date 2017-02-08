package main

import (
	"fmt"

	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/types"
)

func init() {
	commands.RegisterTxSubcommand(CounterTxCmd)
	commands.RegisterStartPlugin("counter", func() types.Plugin { return counter.New() })
}

var (
	ValidFlag = cli.BoolFlag{
		Name:  "valid",
		Usage: "Set valid field in CounterTx",
	}

	CounterTxCmd = cli.Command{
		Name:  "counter",
		Usage: "Create, sign, and broadcast a transaction to the counter plugin",
		Action: func(c *cli.Context) error {
			return cmdCounterTx(c)
		},
		Flags: append(commands.TxFlags, ValidFlag),
	}
)

func cmdCounterTx(c *cli.Context) error {
	valid := c.Bool("valid")

	counterTx := counter.CounterTx{
		Valid: valid,
		Fee: types.Coins{
			{
				Denom:  c.String("coin"),
				Amount: int64(c.Int("fee")),
			},
		},
	}

	fmt.Println("CounterTx:", string(wire.JSONBytes(counterTx)))

	data := wire.BinaryBytes(counterTx)
	name := "counter"

	return commands.AppTx(c, name, data)
}
