package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
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
func GetSignBatchCommand(clientCtx client.Context) *cobra.Command {
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
		RunE:   makeSignBatchCmd(clientCtx),
		Args:   cobra.ExactArgs(1),
	}

	cmd.Flags().String(flagMultisig, "", "Address of the multisig account on behalf of which the transaction shall be signed")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().Bool(flagSigOnly, true, "Print only the generated signature, then exit")
	cmd.MarkFlagRequired(flags.FlagFrom)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func makeSignBatchCmd(clientCtx client.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		inBuf := bufio.NewReader(cmd.InOrStdin())
		clientCtx = clientCtx.InitWithInput(inBuf)
		txFactory := tx.NewFactoryCLI(clientCtx, cmd.Flags())

		txGen := clientCtx.TxGenerator
		txBuilder := txGen.NewTxBuilder()
		var err error
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

		for sequence := txFactory.Sequence(); scanner.Scan(); sequence++ {
			var stdTx types.StdTx

			unsignedStdTx := scanner.StdTx()
			txFactory = txFactory.WithSequence(sequence)

			if multisigAddr.Empty() {
				homeDir, _ := cmd.Flags().GetString(flags.FlagFrom)
				err = authclient.SignStdTx(txFactory, clientCtx, homeDir, txBuilder, false, true)
			} else {
				err = authclient.SignStdTxWithSignerAddress(txFactory, clientCtx, multisigAddr, clientCtx.GetFromName(), txBuilder, true)
			}

			if err != nil {
				return err
			}

			json, err := newGetSignatureJSON(clientCtx.JSONMarshaler, txGen, txBuilder, stdTx, generateSignatureOnly)
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
func GetSignCommand(clientCtx client.Context) *cobra.Command {
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
		RunE:   makeSignCmd(clientCtx),
		Args:   cobra.ExactArgs(1),
	}

	cmd.Flags().String(flagMultisig, "", "Address of the multisig account on behalf of which the transaction shall be signed")
	cmd.Flags().Bool(flagAppend, true, "Append the signature to the existing ones. If disabled, old signatures would be overwritten. Ignored if --multisig is on")
	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().String(flags.FlagHome, "", "The application home directory")
	cmd.MarkFlagRequired(flags.FlagFrom)
	flags.AddTxFlagsToCmd(cmd)

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

func makeSignCmd(clientCtx client.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, txF, newTx, err := readTxAndInitContexts(clientCtx, cmd, args[0])
		if err != nil {
			return err
		}
		txGen := clientCtx.TxGenerator
		txBuilder := txGen.NewTxBuilder()
		// if --signature-only is on, then override --append
		generateSignatureOnly, _ := cmd.Flags().GetBool(flagSigOnly)
		multisigAddrStr, _ := cmd.Flags().GetString(flagMultisig)

		if multisigAddrStr != "" {
			var multisigAddr sdk.AccAddress

			multisigAddr, err = sdk.AccAddressFromBech32(multisigAddrStr)
			if err != nil {
				return err
			}

			err = authclient.SignStdTxWithSignerAddress(
				txF, clientCtx, multisigAddr, clientCtx.GetFromName(), txBuilder, clientCtx.Offline,
			)
			generateSignatureOnly = true
		} else {
			append, _ := cmd.Flags().GetBool(flagAppend)
			appendSig := append && !generateSignatureOnly
			if appendSig {
				err = authclient.SignStdTx(txF, clientCtx, clientCtx.GetFromName(), txBuilder, appendSig, clientCtx.Offline)
			}
		}

		if err != nil {
			return err
		}

		json, err := newGetSignatureJSON(clientCtx.JSONMarshaler, txGen, txBuilder, newTx, generateSignatureOnly)
		if err != nil {
			return err
		}

		outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
		if outputDoc == "" {
			fmt.Printf("%s\n", json)
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

func getSignatureJSON(cdc *codec.Codec, newTx types.StdTx, generateSignatureOnly bool) ([]byte, error) {
	if generateSignatureOnly {
		return cdc.MarshalJSON(newTx.Signatures[0])
	}

	return cdc.MarshalJSON(newTx)
}

func newGetSignatureJSON(cdc codec.JSONMarshaler, txGen client.TxGenerator, txBldr client.TxBuilder, newTx sdk.Tx, generateSignatureOnly bool) ([]byte, error) {
	if generateSignatureOnly {
		return cdc.MarshalJSON(txBldr.GetTx().GetSignatures())
	}

	return txGen.TxEncoder()(newTx)
}
