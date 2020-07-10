package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	flagMultisig = "multisig"
	flagAppend   = "append"
	flagSigOnly  = "signature-only"
)

// GetSignBatchCommand returns the transaction sign-batch command.
func GetSignBatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-batch [file]",
		Short: "Sign transaction batch files",
		Long: `Sign batch files of transactions generated with --generate-only.
The command processes list of transactions from file (one StdTx each line), generate
signed transactions or signatures and print their JSON encoding, delimited by '\n'.
As the signatures are generated, the command updates the sequence number accordingly.

If the flag --signature-only flag is set, it will output a JSON representation
of the generated signature only.

The --offline flag makes sure that the client will not reach out to full node.
As a result, the account and the sequence number queries will not be performed and
it is required to set such parameters manually. Note, invalid values will cause
the transaction to fail. The sequence will be incremented automatically for each
transaction that is signed.

The --multisig=<multisig_key> flag generates a signature on behalf of a multisig
account key. It implies --signature-only.
`,
		PreRun: preSignCmd,
		RunE:   makeSignBatchCmd(),
		Args:   cobra.ExactArgs(1),
	}

	cmd.Flags().String(
		flagMultisig, "",
		"Address of the multisig account on behalf of which the transaction shall be signed",
	)
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().Bool(flagSigOnly, true, "Print only the generated signature, then exit")
	cmd.Flags().String(flags.FlagChainID, "", "network chain ID")
	cmd = flags.PostCommands(cmd)[0]
	cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func makeSignBatchCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd)

		txBldr, err := types.NewTxBuilderFromFlags(bufio.NewReader(cmd.InOrStdin()), cmd.Flags(), clientCtx.HomeDir)
		if err != nil {
			return err
		}

		generateSignatureOnly, _ := cmd.Flags().GetBool(flagSigOnly)

		var (
			multisigAddr sdk.AccAddress
			infile       = os.Stdin
		)

		// validate multisig address if there's any
		if ms, _ := cmd.Flags().GetString(flagMultisig); ms != "" {
			multisigAddr, err = sdk.AccAddressFromBech32(ms)
			if err != nil {
				return err
			}
		}

		// prepare output document
		closeFunc, err := setOutputFile(cmd)
		if err != nil {
			return err
		}

		defer closeFunc()
		clientCtx.WithOutput(cmd.OutOrStdout())

		if args[0] != "-" {
			infile, err = os.Open(args[0])
			if err != nil {
				return err
			}
		}

		scanner := authclient.NewBatchScanner(clientCtx.JSONMarshaler, infile)

		for sequence := txBldr.Sequence(); scanner.Scan(); sequence++ {
			var stdTx types.StdTx

			unsignedStdTx := scanner.StdTx()
			txBldr = txBldr.WithSequence(sequence)

			if multisigAddr.Empty() {
				from, _ := cmd.Flags().GetString(flags.FlagFrom)
				_, fromName, err := client.GetFromFields(txBldr.Keybase(), from, clientCtx.GenerateOnly)
				if err != nil {
					return fmt.Errorf("error getting account from keybase: %w", err)
				}

				stdTx, err = authclient.SignStdTx(txBldr, clientCtx, fromName, unsignedStdTx, false, true)
				if err != nil {
					return err
				}
			} else {
				stdTx, err = authclient.SignStdTxWithSignerAddress(txBldr, clientCtx, multisigAddr, clientCtx.GetFromName(), unsignedStdTx, true)
				if err != nil {
					return err
				}
			}

			json, err := getSignatureJSON(clientCtx.JSONMarshaler, stdTx, generateSignatureOnly)
			if err != nil {
				return err
			}

			cmd.Printf("%s\n", json)
		}

		if err := scanner.UnmarshalErr(); err != nil {
			return err
		}

		return scanner.Err()
	}
}

func setOutputFile(cmd *cobra.Command) (func(), error) {
	outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
	if outputDoc == "" {
		cmd.SetOut(cmd.OutOrStdout())
		return func() {}, nil
	}

	fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return func() {}, err
	}

	cmd.SetOut(fp)

	return func() { fp.Close() }, nil
}

// GetSignCommand returns the transaction sign command.
func GetSignCommand() *cobra.Command {
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
		RunE:   makeSignCmd(),
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
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().String(flags.FlagHome, "", "The application home directory")
	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")
	cmd = flags.PostCommands(cmd)[0]
	cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func preSignCmd(cmd *cobra.Command, _ []string) {
	// Conditionally mark the account and sequence numbers required as no RPC
	// query will be done.
	if offline, _ := cmd.Flags().GetBool(flags.FlagOffline); offline {
		cmd.MarkFlagRequired(flags.FlagAccountNumber)
		cmd.MarkFlagRequired(flags.FlagSequence)
	}
}

func makeSignCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd)

		clientCtx, txBldr, tx, err := readStdTxAndInitContexts(clientCtx, cmd, args[0])
		if err != nil {
			return err
		}
		stdTx := tx.(types.StdTx)

		// if --signature-only is on, then override --append
		var newTx types.StdTx
		generateSignatureOnly, _ := cmd.Flags().GetBool(flagSigOnly)
		multisigAddrStr, _ := cmd.Flags().GetString(flagMultisig)

		if multisigAddrStr != "" {
			var multisigAddr sdk.AccAddress

			multisigAddr, err = sdk.AccAddressFromBech32(multisigAddrStr)
			if err != nil {
				return err
			}
			newTx, err = authclient.SignStdTxWithSignerAddress(
				txBldr, clientCtx, multisigAddr, clientCtx.GetFromName(), stdTx, clientCtx.Offline,
			)
			generateSignatureOnly = true
		} else {
			append, _ := cmd.Flags().GetBool(flagAppend)
			appendSig := append && !generateSignatureOnly
			newTx, err = authclient.SignStdTx(txBldr, clientCtx, clientCtx.GetFromName(), stdTx, appendSig, clientCtx.Offline)
		}
		if err != nil {
			return err
		}

		json, err := getSignatureJSON(clientCtx.JSONMarshaler, newTx, generateSignatureOnly)
		if err != nil {
			return err
		}

		outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
		if outputDoc == "" {
			cmd.Printf("%s\n", json)
			return nil
		}

		fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer fp.Close()

		fmt.Fprintf(fp, "%s\n", json)
		return nil
	}
}

func getSignatureJSON(cdc codec.JSONMarshaler, newTx types.StdTx, generateSignatureOnly bool) ([]byte, error) {
	if generateSignatureOnly {
		return cdc.MarshalJSON(newTx.Signatures[0])
	}

	return cdc.MarshalJSON(newTx)
}
