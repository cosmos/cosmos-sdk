package cli

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client"
)

// GetBroadcastCommand returns the tx broadcast command.
func GetBroadcastCommand(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast [file_path]",
		Short: "Broadcast transactions generated offline",
		Long: strings.TrimSpace(`Broadcast transactions created with the --generate-only
flag and signed with the sign command. Read a transaction from [file_path] and
broadcast it to a node. If you supply a dash (-) argument in place of an input
filename, the command reads from standard input.

$ <appcli> tx broadcast ./mytxn.json
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			if cliCtx.Offline {
				return errors.New("cannot broadcast tx during offline mode")
			}

			stdTx, err := client.ReadStdTxFromFile(cliCtx.Codec, args[0])
			if err != nil {
				return err
			}

			txBytes, err := cliCtx.Codec.MarshalBinaryBare(stdTx)
			if err != nil {
				return err
			}

			res, err := cliCtx.BroadcastTx(txBytes)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(res)
		},
	}

	return flags.PostCommands(cmd)[0]
}
