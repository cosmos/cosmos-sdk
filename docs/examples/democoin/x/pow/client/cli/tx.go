package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/pow"

	"github.com/spf13/cobra"
)

// command to mine some pow!
func MineCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mine [difficulty] [count] [nonce] [solution]",
		Short: "Mine some coins with proof-of-work!",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

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

			msg := pow.NewMsgMine(cliCtx.FromAddr(), difficulty, count, nonce, []byte(args[3]))
			return cliCtx.MessageOutput(msg)
		},
	}
}
