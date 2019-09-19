package cli

import (
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// txDecodeRespTx implements a simple Stringer wrapper for a decoded StdTx.
type txDecodeRespTx authtypes.StdTx

func (tx txDecodeRespTx) String() string {
	txString := fmt.Sprintf(`
	Msgs: 			%s
	Fee:  			%v
	Signatures: %s
	Memo: 			%s
	`, tx.Msgs, tx.Fee,
		tx.Signatures, tx.Memo)
	return txString
}

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

			return cliCtx.PrintOutput(txDecodeRespTx(stdTx))
		},
	}

	return client.PostCommands(cmd)[0]
}
