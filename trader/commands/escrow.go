package commands

import (
	"github.com/tendermint/basecoin-examples/trader/escrow"
	bcmd "github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/types"
	"github.com/urfave/cli"
)

const EscrowName = "escrow"

// var (
// 	MintToFlag = cli.StringFlag{
// 		Name:  "mintto",
// 		Usage: "Where to send the newly minted coins",
// 	}
// 	MintAmountFlag = cli.IntFlag{
// 		Name:  "mint",
// 		Usage: "Amount of coins to mint",
// 	}
// 	MintCoinFlag = cli.StringFlag{
// 		Name:  "mintcoin",
// 		Value: "blank",
// 		Usage: "Specify a coin denomination to mint",
// 	}
// )

var (
	EscrowTxCmd = cli.Command{
		Name:  "escrow",
		Usage: "Put a payment in escrow until there is agreement",
		Action: func(c *cli.Context) error {
			return cmdEscrowTx(c)
		},
		Flags: []cli.Flag{
		// MintToFlag,
		// MintAmountFlag,
		// MintCoinFlag,
		},
	}

	EscrowPluginFlag = cli.BoolFlag{
		Name:  "escrow-plugin",
		Usage: "Enable the escrow plugin",
	}
)

func init() {
	bcmd.RegisterTxPlugin(EscrowTxCmd)
	bcmd.RegisterStartPlugin(EscrowPluginFlag,
		func() types.Plugin { return escrow.New(EscrowName) })
}

func cmdEscrowTx(c *cli.Context) error {
	// toHex := c.String(MintToFlag.Name)
	// mintAmount := int64(c.Int(MintAmountFlag.Name))
	// mintCoin := c.String(MintCoinFlag.Name)
	parent := c.Parent()

	// // convert destination address to bytes
	// to, err := hex.DecodeString(bcmd.StripHex(toHex))
	// if err != nil {
	// 	return errors.New("To address is invalid hex: " + err.Error())
	// }

	// mintTx := mintcoin.MintTx{
	// 	Winners: []mintcoin.Winner{
	// 		{
	// 			Addr: to,
	// 			Amount: types.Coins{
	// 				{
	// 					Denom:  mintCoin,
	// 					Amount: mintAmount,
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// fmt.Println("MintTx:", string(wire.JSONBytes(mintTx)))
	// data := wire.BinaryBytes(mintTx)

	data := []byte("foo")
	return bcmd.AppTx(parent, EscrowName, data)
}
