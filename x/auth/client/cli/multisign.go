package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

// GetSignCommand returns the sign command
func GetMultiSignCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign [file] [name] [[signature]...]",
		Short: "Generate multisig signatures for transactions generated offline",
		Long: `Sign transactions created with the --generate-only flag that require multisig signatures.

Read signature(s) from [signature] file(s), generate a multisig signature compliant to the
multisig key [name], and attach it to the transaction read from [file]. Example:

   gaiacli multisign transaction.json k1k2k3 k1sig.json k2sig.json k3sig.json

If the flag --signature-only flag is on, it outputs a JSON representation
of the generated signature only.

The --offline flag makes sure that the client will not reach out to an external node.
Thus account number or sequence number lookups will not be performed and it is
recommended to set such parameters manually.
`,
		RunE: makeMultiSignCmd(codec),
		Args: cobra.MinimumNArgs(3),
	}
	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().Bool(flagOffline, false, "Offline mode. Do not query a full node")
	cmd.Flags().String(flagOutfile, "", "The document will be written to the given file instead of STDOUT")

	// Add the flags here and return the command
	return client.PostCommands(cmd)[0]
}

func makeMultiSignCmd(cdc *amino.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		stdTx, err := authclient.ReadStdTxFromFile(cdc, args[0])
		if err != nil {
			return
		}

		keybase, err := keys.NewKeyBaseFromDir(viper.GetString(cli.HomeFlag))
		if err != nil {
			return
		}

		multisigInfo, err := keybase.Get(args[1])
		if err != nil {
			return
		}
		if multisigInfo.GetType() != crkeys.TypeOffline {
			return fmt.Errorf("%q must be of type offline: %s",
				args[1], multisigInfo.GetType())
		}

		multisigPub := multisigInfo.GetPubKey().(multisig.PubKeyMultisigThreshold)
		multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))
		cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
		txBldr := authtxb.NewTxBuilderFromCLI()

		if !viper.GetBool(flagOffline) {
			addr := multisigInfo.GetAddress()
			accnum, err := cliCtx.GetAccountNumber(addr)
			if err != nil {
				return err
			}

			seq, err := cliCtx.GetAccountSequence(addr)
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
			sigBytes := auth.StdSignBytes(
				txBldr.ChainID(), txBldr.AccountNumber(), txBldr.Sequence(),
				stdTx.Fee, stdTx.GetMsgs(), stdTx.GetMemo(),
			)
			if ok := stdSig.PubKey.VerifyBytes(sigBytes, stdSig.Signature); !ok {
				return fmt.Errorf("couldn't verify signature")
			}
			multisigSig.AddSignatureFromPubKey(stdSig.Signature, stdSig.PubKey, multisigPub.PubKeys)
		}

		newStdSig := auth.StdSignature{Signature: cdc.MustMarshalBinaryBare(multisigSig), PubKey: multisigPub}
		newTx := auth.NewStdTx(stdTx.GetMsgs(), stdTx.Fee, []auth.StdSignature{newStdSig}, stdTx.GetMemo())

		sigOnly := viper.GetBool(flagSigOnly)
		var json []byte
		switch {
		case sigOnly && cliCtx.Indent:
			json, err = cdc.MarshalJSONIndent(newTx.Signatures[0], "", "  ")
		case sigOnly && !cliCtx.Indent:
			json, err = cdc.MarshalJSON(newTx.Signatures[0])
		case !sigOnly && cliCtx.Indent:
			json, err = cdc.MarshalJSONIndent(newTx, "", "  ")
		default:
			json, err = cdc.MarshalJSON(newTx)
		}
		if err != nil {
			return err
		}

		if viper.GetString(flagOutfile) == "" {
			fmt.Printf("%s\n", json)
			return
		}

		fp, err := os.OpenFile(
			viper.GetString(flagOutfile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644,
		)
		if err != nil {
			return err
		}
		defer fp.Close()

		fmt.Fprintf(fp, "%s\n", json)

		return
	}
}

func readAndUnmarshalStdSignature(cdc *amino.Codec, filename string) (stdSig auth.StdSignature, err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = cdc.UnmarshalJSON(bytes, &stdSig); err != nil {
		return
	}
	return
}
