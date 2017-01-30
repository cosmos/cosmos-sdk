package commands

import (
	"fmt"

	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"
)

var (
	CounterTxCmd = cli.Command{
		Name:  "counter",
		Usage: "Craft a transaction to the counter plugin",
		Action: func(c *cli.Context) error {
			return cmdCounterTx(c)
		},
		Flags: []cli.Flag{
			ValidFlag,
		},
	}

	CounterPluginFlag = cli.BoolFlag{
		Name:  "counter-plugin",
		Usage: "Enable the counter plugin",
	}
)

func init() {
	RegisterTxPlugin(CounterTxCmd)
	RegisterStartPlugin(CounterPluginFlag,
		func() types.Plugin { return counter.New("counter") })
}

func cmdCounterTx(c *cli.Context) error {
	valid := c.Bool("valid")
	parent := c.Parent()

	counterTx := counter.CounterTx{
		Valid: valid,
		Fee: types.Coins{
			{
				Denom:  parent.String("coin"),
				Amount: int64(parent.Int("fee")),
			},
		},
	}

	fmt.Println("CounterTx:", string(wire.JSONBytes(counterTx)))

	data := wire.BinaryBytes(counterTx)
	name := "counter"

	return AppTx(parent, name, data)
}
