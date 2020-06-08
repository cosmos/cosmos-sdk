package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	types2 "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetSignCommand returns the sign command
func GetMultiSignCommand(ctx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign [file] [name] [[signature]...]",
		Short: "Generate multisig signatures for transactions generated offline",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Sign transactions created with the --generate-only flag that require multisig signatures.

Read signature(s) from [signature] file(s), generate a multisig signature compliant to the
multisig key [name], and attach it to the transaction read from [file].

Example:
$ %s multisign transaction.json k1k2k3 k1sig.json k2sig.json k3sig.json

If the flag --signature-only flag is on, it outputs a JSON representation
of the generated signature only.

The --offline flag makes sure that the client will not reach out to an external node.
Thus account number or sequence number lookups will not be performed and it is
recommended to set such parameters manually.
`,
				version.ClientName,
			),
		),
		RunE: makeMultiSignCmd(ctx),
		Args: cobra.MinimumNArgs(3),
	}

	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")

	// Add the flags here and return the command
	return flags.PostCommands(cmd)[0]
}

func makeMultiSignCmd(clientCtx client.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		stdTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
		if err != nil {
			return
		}

		inBuf := bufio.NewReader(cmd.InOrStdin())
		kb, err := keyring.New(sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), inBuf)
		if err != nil {
			return
		}

		multisigInfo, err := kb.Key(args[1])
		if err != nil {
			return
		}
		if multisigInfo.GetType() != keyring.TypeMulti {
			return fmt.Errorf("%q must be of type %s: %s", args[1], keyring.TypeMulti, multisigInfo.GetType())
		}

		multisigPub := multisigInfo.GetPubKey().(multisig.MultisigPubKey)
		multisigSig := multisig.NewMultisig(len(multisigPub.GetPubKeys()))
		clientCtx := clientCtx.InitWithInput(inBuf)
		txFactory := tx.NewFactoryFromCLI(inBuf)

		if !clientCtx.Offline {
			accnum, seq, err := types.NewAccountRetriever(authclient.Codec).GetAccountNumberSequence(clientCtx, multisigInfo.GetAddress())
			if err != nil {
				return err
			}

			txFactory = txFactory.WithAccountNumber(accnum).WithSequence(seq)
		}

		feeTx := stdTx.(types2.FeeTx)
		fee := types.StdFee{
			Amount: feeTx.GetFee(),
			Gas:    feeTx.GetGas(),
		}

		memoTx := stdTx.(types2.TxWithMemo)

		// read each signature and add it to the multisig if valid
		for i := 2; i < len(args); i++ {
			stdSig, err := readAndUnmarshalStdSignature(clientCtx, args[i])
			if err != nil {
				return err
			}

			signingData := types2.SigningData{
				PublicKey:       stdSig.PubKey,
				ChainID:         txFactory.ChainID(),
				AccountNumber:   txFactory.AccountNumber(),
				AccountSequence: txFactory.Sequence(),
			}

			// Validate each signature
			if ok := signing.VerifySignature(signingData, stdSig.Data, stdTx, clientCtx.TxGenerator.SignModeHandler()); !ok {
				return fmt.Errorf("couldn't verify signature")
			}
			if err := multisig.AddSignatureFromPubKey(multisigSig, stdSig.Data, stdSig.PubKey, multisigPub.GetPubKeys()); err != nil {
				return err
			}
		}

		newStdSig := types.StdSignature{Signature: clientCtx.Codec.MustMarshalBinaryBare(multisigSig), PubKey: multisigPub.Bytes()} //nolint:staticcheck
		newTx := types.NewStdTx(stdTx.GetMsgs(), fee, []types.StdSignature{newStdSig}, memoTx.GetMemo())                            //nolint:staticcheck

		txBldr, err := clientCtx.TxGenerator.WrapTxBuilder(stdTx)
		if err != nil {
			return err
		}

		sigBldr := client.SignatureBuilder{
			PubKey: multisigPub,
			Data:   multisigSig,
		}

		err = txBldr.SetSignatures(sigBldr)

		if err != nil {
			return err
		}

		sigOnly := viper.GetBool(flagSigOnly)
		var json []byte
		switch {
		case sigOnly && clientCtx.Indent:
			json, err = getSignatureBuilderJSON(clientCtx.JSONMarshaler, sigBldr, true)
		case sigOnly && !clientCtx.Indent:
			json, err = getSignatureBuilderJSON(clientCtx.JSONMarshaler, sigBldr, false)
		case !sigOnly && clientCtx.Indent:
			json, err = codec.MarshalJSONIndent(clientCtx.JSONMarshaler, newTx)
		default:
			json, err = clientCtx.JSONMarshaler.MarshalJSON(newTx)
		}
		if err != nil {
			return err
		}

		if viper.GetString(flags.FlagOutputDocument) == "" {
			fmt.Printf("%s\n", json)
			return
		}

		fp, err := os.OpenFile(
			viper.GetString(flags.FlagOutputDocument), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644,
		)
		if err != nil {
			return err
		}
		defer fp.Close()

		fmt.Fprintf(fp, "%s\n", json)

		return
	}
}

func readAndUnmarshalStdSignature(clientCtx client.Context, filename string) (stdSig client.SignatureBuilder, err error) { //nolint:staticcheck
	//var bytes []byte
	//if bytes, err = ioutil.ReadFile(filename); err != nil {
	//	return
	//}
	//if err = clientCtx.JSONMarshaler.UnmarshalJSON(bytes, &stdSig); err != nil {
	//	return
	//}
	//return
	panic("TODO: json signatures")
}
