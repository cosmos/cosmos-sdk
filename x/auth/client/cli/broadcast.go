package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			stdTx, err := utils.ReadStdTxFromFile(cliCtx.Codec, args[0])
			if err != nil {
				return
			}

			txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(stdTx)
			if err != nil {
				return
			}

			res, err := cliCtx.BroadcastTx(txBytes)
			cliCtx.PrintOutput(res) // nolint:errcheck

			return err
		},
	}

	return flags.PostCommands(cmd)[0]
}
