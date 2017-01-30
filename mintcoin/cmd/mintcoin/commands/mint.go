package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/basecoin-examples/mintcoin"
	bcmd "github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"
)

var (
	MintAmountFlag = cli.IntFlag{
		Name:  "mint",
		Value: 0,
		Usage: "Amount of coins to mint",
	}
	MintCoinFlag = cli.StringFlag{
		Name:  "mintcoin",
		Value: "blank",
		Usage: "Specify a coin denomination to mint",
	}
)

var (
	MintTxCmd = cli.Command{
		Name:  "mint",
		Usage: "Craft a transaction to mint some more currency",
		Action: func(c *cli.Context) error {
			return cmdMintTx(c)
		},
		Flags: []cli.Flag{
			bcmd.ToFlag,
			MintAmountFlag,
			MintCoinFlag,
		},
	}

	MintPluginFlag = cli.BoolFlag{
		Name:  "mint-plugin",
		Usage: "Enable the mintcoin plugin",
	}
)

func init() {
	bcmd.RegisterTxPlugin(MintTxCmd)
	bcmd.RegisterStartPlugin(MintPluginFlag,
		func() types.Plugin { return mintcoin.New("mint") })
}

func cmdMintTx(c *cli.Context) error {
	// valid := c.Bool("valid")
	toHex := c.String("to")
	mintAmount := int64(c.Int("mint"))
	mintCoin := c.String("mintcoin")
	parent := c.Parent()

	// convert destination address to bytes
	to, err := hex.DecodeString(bcmd.StripHex(toHex))
	if err != nil {
		return errors.New("To address is invalid hex: " + err.Error())
	}

	mintTx := mintcoin.MintTx{
		Winners: []mintcoin.Winner{
			{
				Addr: to,
				Amount: types.Coins{
					{
						Denom:  mintCoin,
						Amount: mintAmount,
					},
				},
			},
		},
	}
	fmt.Println("MintTx:", string(wire.JSONBytes(mintTx)))

	data := wire.BinaryBytes(mintTx)
	name := "mint"

	return bcmd.AppTx(parent, name, data)
}
