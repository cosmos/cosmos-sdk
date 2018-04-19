package commands

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/examples/democoin/x/pow"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"
)

// command to mine some pow!
func MineCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mine [difficulty] [count] [nonce] [solution]",
		Short: "Mine some coins with proof-of-work!",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 4 {
				return errors.New("You must provide a difficulty, a count, a solution, and a nonce (in that order)")
			}

			// get from address and parse arguments

			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			difficulty, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}
			count, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}
			nonce, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return err
			}

			solution := []byte(args[3])

			msg := pow.NewMsgMine(from, difficulty, count, nonce, solution)

			// get account name
			name := ctx.FromAddressName

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.EnsureSignBuildBroadcast(name, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}
