package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
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
		Use:   "sign [[file]...]",
		Short: "Sign transactions generated offline",
		Long: `Sign transactions created with the --generate-only flag.
Read a transactions from file, sign them and print their JSON encoding to STDOUT.

If the flag --signature-only flag is set, the command will outputs a JSON
representation of the generated signatures only.

The --offline flag makes sure that the client will not reach out to full node.
As a result, the account and sequence number queries will not be performed and
it is required to set such parameters manually. Note that invalid values will
cause the transaction to fail when they are broadcast eventually.

The --multisig=<multisig_key> flag generates signatures on behalf of a multisig account
key. It implies --signature-only. Full multisig signed transactions may eventually
be generated via the 'multisign' command.
`,
		PreRun: preSignCmd,
		RunE:   makeSignCmd(codec),
		Args:   cobra.MinimumNArgs(1),
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
		var (
			err     error
			newTx   types.StdTx
			outfile *os.File
		)

		inBuf := bufio.NewReader(cmd.InOrStdin())
		cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
		txBldr := types.NewTxBuilderFromCLI(inBuf)
		generateSignatureOnly := viper.GetBool(flagSigOnly) // if --signature-only is on, then override --append
		multisigAddrStr := viper.GetString(flagMultisig)
		outfileName := viper.GetString(flagOutfile)

		if outfileName != "" {
			outfile, err = os.OpenFile(outfileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer outfile.Close()
		}

		for _, filename := range args {
			stdTx, err := client.ReadStdTxFromFile(cdc, filename)
			if err != nil {
				return fmt.Errorf("couldn't read %s: %v", filename, err)
			}

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

			if outfileName != "" {
				fmt.Fprintf(outfile, "%s\n", json)
				continue
			}

			fmt.Printf("%s\n", json)
		}

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
