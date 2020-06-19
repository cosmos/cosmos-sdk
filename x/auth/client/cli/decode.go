package cli

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

const flagHex = "hex"

// GetDecodeCommand returns the decode command to take serialized bytes
// and turn it into a JSONified transaction.
func GetDecodeCommand(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [amino-byte-string]",
		Short: "Decode an binary encoded transaction string.",
		Args:  cobra.ExactArgs(1),
		RunE:  runDecodeTxString(clientCtx),
	}

	cmd.Flags().BoolP(flagHex, "x", false, "Treat input as hexadecimal instead of base64")
	return flags.PostCommands(cmd)[0]
}

func runDecodeTxString(clientCtx client.Context) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		clientCtx = clientCtx.Init().WithOutput(cmd.OutOrStdout())
		var txBytes []byte

		if viper.GetBool(flagHex) {
			txBytes, err = hex.DecodeString(args[0])
		} else {
			txBytes, err = base64.StdEncoding.DecodeString(args[0])
		}
		if err != nil {
			return err
		}

		tx, err := clientCtx.TxGenerator.TxDecoder()(txBytes)
		if err != nil {
			return err
		}

		return clientCtx.PrintOutput(tx)
	}
}
