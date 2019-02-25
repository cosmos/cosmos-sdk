package cli

import (
	"strings"

	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

// GetSignCommand returns the sign command
func GetBroadcastCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast [file_path]",
		Short: "Broadcast transactions generated offline",
		Long: strings.TrimSpace(`Broadcast transactions created with the --generate-only flag and signed with the sign command.
Read a transaction from [file_path] and broadcast it to a node. If you supply a dash (-) argument
in place of an input filename, the command reads from standard input.

$ gaiacli tx broadcast ./mytxn.json
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec)
			stdTx, err := authclient.ReadStdTxFromFile(cliCtx.Codec, args[0])
			if err != nil {
				return
			}

			txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(stdTx)
			if err != nil {
				return
			}

			res, err := cliCtx.BroadcastTx(txBytes)
			cliCtx.PrintOutput(res)
			return err
		},
	}

	return client.PostCommands(cmd)[0]
}
