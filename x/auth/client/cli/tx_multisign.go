package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetSignCommand returns the sign command
func GetMultiSignCommand(clientCtx client.Context) *cobra.Command {
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
				version.AppName,
			),
		),
		RunE: makeMultiSignCmd(clientCtx),
		Args: cobra.MinimumNArgs(3),
	}

	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")

	return flags.PostCommands(cmd)[0]
}

func makeMultiSignCmd(clientCtx client.Context) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		clientCtx = clientCtx.Init()
		cdc := clientCtx.Codec
		tx, err := authclient.ReadTxFromFile(clientCtx, args[0])
		stdTx := tx.(types.StdTx)
		if err != nil {
			return
		}

		backend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
		homeDir, _ := cmd.Flags().GetString(flags.FlagHome)

		inBuf := bufio.NewReader(cmd.InOrStdin())
		kb, err := keyring.New(sdk.KeyringServiceName(), backend, homeDir, inBuf)
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

		multisigPub := multisigInfo.GetPubKey().(multisig.PubKeyMultisigThreshold)
		multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))
		txBldr, err := types.NewTxBuilderFromFlags(inBuf, cmd.Flags(), homeDir)
		if err != nil {
			return errors.Wrap(err, "error creating tx builder from flags")
		}

		if !clientCtx.Offline {
			accnum, seq, err := types.NewAccountRetriever(authclient.Codec).GetAccountNumberSequence(clientCtx, multisigInfo.GetAddress())
			if err != nil {
				return err
			}

			txBldr = txBldr.WithAccountNumber(accnum).WithSequence(seq)
		}

		// read each signature and add it to the multisig if valid
		for i := 2; i < len(args); i++ {
			stdSig, err := readAndUnmarshalStdSignature(cdc, args[i])
			if err != nil {
				return err
			}

			// Validate each signature
			sigBytes := types.StdSignBytes(
				txBldr.ChainID(), txBldr.AccountNumber(), txBldr.Sequence(),
				stdTx.Fee, stdTx.GetMsgs(), stdTx.GetMemo(),
			)
			if ok := stdSig.GetPubKey().VerifyBytes(sigBytes, stdSig.Signature); !ok {
				return fmt.Errorf("couldn't verify signature")
			}

			sigV2, err := types.StdSignatureToSignatureV2(cdc, stdSig)
			if err != nil {
				return nil
			}

			if err := multisig.AddSignatureV2(multisigSig, sigV2, multisigPub.PubKeys); err != nil {
				return err
			}
		}

		sigBz, err := types.SignatureDataToAminoSignature(cdc, multisigSig)
		if err != nil {
			return err
		}

		newStdSig := types.StdSignature{Signature: sigBz, PubKey: multisigPub.Bytes()}                        //nolint:staticcheck
		newTx := types.NewStdTx(stdTx.GetMsgs(), stdTx.Fee, []types.StdSignature{newStdSig}, stdTx.GetMemo()) //nolint:staticcheck

		var json []byte

		sigOnly, _ := cmd.Flags().GetBool(flagSigOnly)
		if sigOnly {
			json, err = cdc.MarshalJSON(newTx.Signatures[0])
		} else {
			json, err = cdc.MarshalJSON(newTx)
		}

		if err != nil {
			return err
		}

		outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
		if outputDoc == "" {
			fmt.Printf("%s\n", json)
			return
		}

		fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer fp.Close()

		fmt.Fprintf(fp, "%s\n", json)
		return
	}
}

func readAndUnmarshalStdSignature(cdc *codec.Codec, filename string) (stdSig types.StdSignature, err error) { //nolint:staticcheck
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = cdc.UnmarshalJSON(bytes, &stdSig); err != nil {
		return
	}
	return
}
