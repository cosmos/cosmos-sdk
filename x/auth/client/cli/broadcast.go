package cli

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
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
			clientCtx := client.NewContext().WithCodec(cdc)

			if clientCtx.Offline {
				return errors.New("cannot broadcast tx during offline mode")
			}

			stdTx, err := authclient.ReadStdTxFromFile(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			txBytes, err := clientCtx.Codec.MarshalBinaryBare(stdTx)
			if err != nil {
				return err
			}

			res, err := clientCtx.BroadcastTx(txBytes)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	return flags.PostCommands(cmd)[0]
}
