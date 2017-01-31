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

const MintName = "mint"

var (
	MintToFlag = cli.StringFlag{
		Name:  "mintto",
		Usage: "Where to send the newly minted coins",
	}
	MintAmountFlag = cli.IntFlag{
		Name:  "mint",
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
			MintToFlag,
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
		func() types.Plugin { return mintcoin.New(MintName) })
}

func cmdMintTx(c *cli.Context) error {
	toHex := c.String(MintToFlag.Name)
	mintAmount := int64(c.Int(MintAmountFlag.Name))
	mintCoin := c.String(MintCoinFlag.Name)
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

	return bcmd.AppTx(parent, MintName, data)
}
