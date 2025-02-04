package cli

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

const flagHex = "hex"

// GetDecodeCommand returns the decode command to take serialized bytes and turn
// it into a JSON-encoded transaction.
func GetDecodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode <protobuf-byte-string>",
		Short: "Decode a binary encoded transaction string",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)
			var txBytes []byte

			if useHex, _ := cmd.Flags().GetBool(flagHex); useHex {
				txBytes, err = hex.DecodeString(args[0])
			} else {
				txBytes, err = base64.StdEncoding.DecodeString(args[0])
			}
			if err != nil {
				return err
			}

			json, err := decodeTxAndGetJSON(clientCtx, txBytes)
			if err != nil {
				return err
			}

			return clientCtx.PrintBytes(json)
		},
	}

	cmd.Flags().BoolP(flagHex, "x", false, "Treat input as hexadecimal instead of base64")
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.Flags().MarkHidden(flags.FlagOutput) // decoding makes sense to output only json

	return cmd
}

func decodeTxAndGetJSON(clientCtx client.Context, txBytes []byte) ([]byte, error) {
	// First try decoding with TxDecoder
	tx, err := clientCtx.TxConfig.TxDecoder()(txBytes)
	if err == nil {
		return clientCtx.TxConfig.TxJSONEncoder()(tx)
	}

	// Fallback to direct unmarshaling
	var sdkTx sdktx.Tx
	if err := clientCtx.Codec.Unmarshal(txBytes, &sdkTx); err != nil {
		return nil, err
	}

	return clientCtx.Codec.MarshalJSON(&sdkTx)
}
