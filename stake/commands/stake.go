package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/basecoin-examples/stake"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"github.com/urfave/cli"
)

var bondCmd = cli.Command{
	Name:   "bond",
	Usage:  "Bond some coins to give voting power to a validator",
	Action: cmdBond,
	Flags: append(bcmd.TxFlags,
		cli.StringFlag{
			Name:  "validator",
			Usage: "Validator's public key",
		},
		cli.IntFlag{
			Name:  "amount",
			Usage: "Amount of coins",
		},
	),
}

func init() {
	bcmd.RegisterTxSubcommand(bondCmd)
	bcmd.RegisterStartPlugin("stake",
		func() types.Plugin {
			return stake.New(stake.Params{
				UnbondingPeriod: 100,
				TokenDenom:      "atom",
			})
		},
	)
}

func cmdBond(c *cli.Context) error {
	validatorHex := c.String("validator")

	// convert validator pubkey to bytes
	validator, err := hex.DecodeString(bcmd.StripHex(validatorHex))
	if err != nil {
		return errors.New("Validator is invalid hex: " + err.Error())
	}

	bondTx := stake.BondTx{ValidatorPubKey: validator}
	fmt.Println("BondTx:", string(wire.JSONBytes(bondTx)))
	bytes := wire.BinaryBytes(bondTx)
	return bcmd.AppTx(c, "stake", bytes)
}
