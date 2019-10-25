package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

// DecodeTxCmd - returns the command to decode a tx from hex or base64
func DecodeTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "tx",
		Short: "Decode a tx from hex or base64",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			txString := args[0]

			// try hex, then base64
			txBytes, err := hex.DecodeString(txString)
			if err != nil {
				var err2 error
				txBytes, err2 = base64.StdEncoding.DecodeString(txString)
				if err2 != nil {
					return fmt.Errorf(`expected hex or base64. Got errors:
				hex: %v,
				base64: %v
				`, err, err2)
				}
			}

			var tx = types.StdTx{}

			err = cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(tx)
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer([]byte{})
			err = json.Indent(buf, bz, "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(buf.String())
			return nil
		},
	}
}
