package cli

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// txDecodeRespStr implements a simple Stringer wrapper for a decoded tx.
type txDecodeRespTx authtypes.StdTx

func (tx txDecodeRespTx) String() string {
	return tx.String()
}

// GetDecodeCommand returns the decode command to take Amino-serialized bytes and turn it into
// a JSONified transaction
func GetDecodeCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [amino-byte-string]",
		Short: "Decode an amino-encoded transaction string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliCtx := context.NewCLIContext().WithCodec(codec)

			txBytesBase64 := args[0]

			txBytes, err := base64.StdEncoding.DecodeString(txBytesBase64)
			if err != nil {
				return err
			}

			var stdTx authtypes.StdTx
			err = cliCtx.Codec.UnmarshalBinaryLengthPrefixed(txBytes, &stdTx)
			if err != nil {
				return err
			}

			response := txDecodeRespTx(stdTx)
			_ = cliCtx.PrintOutput(response)

			return nil
		},
	}

	return client.PostCommands(cmd)[0]
}
