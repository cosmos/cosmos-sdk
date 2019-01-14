package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/pow"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/spf13/cobra"
)

// command to mine some pow!
func MineCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mine [difficulty] [count] [nonce] [solution]",
		Short: "Mine some coins with proof-of-work!",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			from, err := cliCtx.GetFromAddress()
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

			// Build and sign the transaction, then broadcast to a Tendermint
			// node.
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
}
