package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/basecoin-examples/mintcoin"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
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
	MintAmountFlag = cli.StringFlag{
		Name:  "mint",
		Usage: "Amount of coins to mint in format <amt><coin>,<amt2><coin2>,...",
	}
)

var (
	MintTxCmd = cli.Command{
		Name:  "mint",
		Usage: "Craft a transaction to mint some more currency",
		Action: func(c *cli.Context) error {
			return cmdMintTx(c)
		},
		Flags: append(bcmd.TxFlags,
			MintToFlag,
			MintAmountFlag),
	}
)

func init() {
	bcmd.RegisterTxSubcommand(MintTxCmd)
	bcmd.RegisterStartPlugin(MintName,
		func() types.Plugin { return mintcoin.New(MintName) })
}

func cmdMintTx(c *cli.Context) error {
	toHex := c.String(MintToFlag.Name)
	mintAmount := c.String(MintAmountFlag.Name)

	// convert destination address to bytes
	to, err := hex.DecodeString(bcmd.StripHex(toHex))
	if err != nil {
		return errors.New("To address is invalid hex: " + err.Error())
	}

	amountCoins, err := bcmd.ParseCoins(mintAmount)
	if err != nil {
		return err
	}

	mintTx := mintcoin.MintTx{
		Credits: []mintcoin.Credit{
			{
				Addr:   to,
				Amount: amountCoins,
			},
		},
	}
	fmt.Println("MintTx:", string(wire.JSONBytes(mintTx)))
	data := wire.BinaryBytes(mintTx)

	return bcmd.AppTx(c, MintName, data)
}
