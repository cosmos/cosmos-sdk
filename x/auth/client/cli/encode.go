package cli

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

// PrintOutput requires a Stringer, so we wrap string
type encodeResp string

func (e encodeResp) String() string {
	return string(e)
}

// GetEncodeCommand returns the encode command to take a JSONified transaction and turn it into
// Amino-serialized bytes
func GetEncodeCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode [file]",
		Short: "encode transactions generated offline",
		Long: `Encode transactions created with the --generate-only flag and signed with the sign command.
Read a transaction from <file>, serialize it to the Amino wire protocol, and output it as base64.
If you supply a dash (-) argument in place of an input filename, the command reads from standard input.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec)

			stdTx, err := authclient.ReadStdTxFromFile(cliCtx.Codec, args[0])
			if err != nil {
				return
			}

			txBytes, err := cliCtx.Codec.MarshalBinaryLengthPrefixed(stdTx)
			if err != nil {
				return err
			}

			// Encode the bytes to base64
			txBytesBase64 := base64.StdEncoding.EncodeToString(txBytes)

			// Write it back
			response := encodeResp(txBytesBase64)
			cliCtx.PrintOutput(response)
			return nil
		},
	}

	return client.PostCommands(cmd)[0]
}
