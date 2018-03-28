package commands

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/x/pow"
	"github.com/cosmos/cosmos-sdk/wire"
)

func MineCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "mine [difficulty] [count] [nonce] [solution]",
		Short: "Mine some coins with proof-of-work!",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 4 {
				return errors.New("You must provide a difficulty, a solution, and a nonce (in that order)")
			}

			// get from address and parse arguments

			from, err := builder.GetFromAddress()
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

			msg := pow.NewMineMsg(from, difficulty, count, nonce, solution)

			// get account name
			name := viper.GetString(client.FlagName)

			// get password
			buf := client.BufferStdin()
			prompt := fmt.Sprintf("Password to sign with '%s':", name)
			passphrase, err := client.GetPassword(prompt, buf)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			res, err := builder.SignBuildBroadcast(name, passphrase, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}
