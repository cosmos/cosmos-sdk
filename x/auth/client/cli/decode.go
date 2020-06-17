package cli

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const flagHex = "hex"

// GetDecodeCommand returns the decode command to take Amino-serialized bytes
// and turn it into a JSONified transaction.
func GetDecodeCommand(codec *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [amino-byte-string]",
		Short: "Decode an amino-encoded transaction string.",
		Args:  cobra.ExactArgs(1),
		RunE:  runDecodeTxString(codec),
	}

	cmd.Flags().BoolP(flagHex, "x", false, "Treat input as hexadecimal instead of base64")
	return flags.PostCommands(cmd)[0]
}

func runDecodeTxString(cdc *codec.Codec) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		clientCtx := client.NewContext().WithCodec(cdc).WithOutput(cmd.OutOrStdout()).WithJSONMarshaler(cdc)
		var txBytes []byte

		if viper.GetBool(flagHex) {
			txBytes, err = hex.DecodeString(args[0])
		} else {
			txBytes, err = base64.StdEncoding.DecodeString(args[0])
		}
		if err != nil {
			return err
		}

		var stdTx authtypes.StdTx
		err = clientCtx.Codec.UnmarshalBinaryBare(txBytes, &stdTx)
		if err != nil {
			return err
		}

		return clientCtx.PrintOutput(stdTx)
	}
}
