package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	flagMultisig = "multisig"
	flagAppend   = "append"
	flagSigOnly  = "signature-only"
	flagOutfile  = "output-document"
)

// GetSignCommand returns the transaction sign command.
func GetSignCommand(codec *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [file]",
		Short: "Sign transactions generated offline",
		Long: `Sign transactions created with the --generate-only flag.
It will read a transaction from [file], sign it, and print its JSON encoding.

If the flag --signature-only flag is set, it will output a JSON representation
of the generated signature only.

The --offline flag makes sure that the client will not reach out to full node.
As a result, the account and sequence number queries will not be performed and
it is required to set such parameters manually. Note, invalid values will cause
the transaction to fail.

The --multisig=<multisig_key> flag generates a signature on behalf of a multisig account
key. It implies --signature-only. Full multisig signed transactions may eventually
be generated via the 'multisign' command.
`,
		PreRun: preSignCmd,
		RunE:   makeSignCmd(codec),
		Args:   cobra.ExactArgs(1),
	}

	cmd.Flags().String(
		flagMultisig, "",
		"Address of the multisig account on behalf of which the transaction shall be signed",
	)
	cmd.Flags().Bool(
		flagAppend, true,
		"Append the signature to the existing ones. If disabled, old signatures would be overwritten. Ignored if --multisig is on",
	)
	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().String(flagOutfile, "", "The document will be written to the given file instead of STDOUT")
	cmd = flags.PostCommands(cmd)[0]
	cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func preSignCmd(cmd *cobra.Command, _ []string) {
	// Conditionally mark the account and sequence numbers required as no RPC
	// query will be done.
	if viper.GetBool(flags.FlagOffline) {
		cmd.MarkFlagRequired(flags.FlagAccountNumber)
		cmd.MarkFlagRequired(flags.FlagSequence)
	}
}

func makeSignCmd(cdc *codec.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cliCtx, txBldr, stdTx, err := readStdTxAndInitContexts(cdc, cmd, args[0])
		if err != nil {
			return err
		}

		// if --signature-only is on, then override --append
		var newTx types.StdTx
		generateSignatureOnly := viper.GetBool(flagSigOnly)
		multisigAddrStr := viper.GetString(flagMultisig)

		if multisigAddrStr != "" {
			var multisigAddr sdk.AccAddress

			multisigAddr, err = sdk.AccAddressFromBech32(multisigAddrStr)
			if err != nil {
				return err
			}
			newTx, err = client.SignStdTxWithSignerAddress(
				txBldr, cliCtx, multisigAddr, cliCtx.GetFromName(), stdTx, cliCtx.Offline,
			)
			generateSignatureOnly = true
		} else {
			appendSig := viper.GetBool(flagAppend) && !generateSignatureOnly
			newTx, err = client.SignStdTx(txBldr, cliCtx, cliCtx.GetFromName(), stdTx, appendSig, cliCtx.Offline)
		}

		if err != nil {
			return err
		}

		json, err := getSignatureJSON(cdc, newTx, cliCtx.Indent, generateSignatureOnly)
		if err != nil {
			return err
		}

		if viper.GetString(flagOutfile) == "" {
			fmt.Printf("%s\n", json)
			return nil
		}

		fp, err := os.OpenFile(
			viper.GetString(flagOutfile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644,
		)
		if err != nil {
			return err
		}

		defer fp.Close()
		fmt.Fprintf(fp, "%s\n", json)

		return nil
	}
}

func getSignatureJSON(cdc *codec.Codec, newTx types.StdTx, indent, generateSignatureOnly bool) ([]byte, error) {
	switch generateSignatureOnly {
	case true:
		switch indent {
		case true:
			return cdc.MarshalJSONIndent(newTx.Signatures[0], "", "  ")

		default:
			return cdc.MarshalJSON(newTx.Signatures[0])
		}
	default:
		switch indent {
		case true:
			return cdc.MarshalJSONIndent(newTx, "", "  ")

		default:
			return cdc.MarshalJSON(newTx)
		}
	}
}
