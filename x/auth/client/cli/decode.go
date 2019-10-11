package cli

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetDecodeCommand returns the decode command to take Amino-serialized bytes
// and turn it into a JSONified transaction.
func GetDecodeCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [amino-byte-string]",
		Short: "Decode an amino-encoded transaction string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec)

			txBytes, err := base64.StdEncoding.DecodeString(args[0])
			if err != nil {
				return err
			}

			var stdTx authtypes.StdTx
			err = cliCtx.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &stdTx)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(stdTx)
		},
	}

	return client.PostCommands(cmd)[0]
}
