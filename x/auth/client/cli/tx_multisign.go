package cli

import (
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/codec"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

//todo: move to flags
const (
	flagNoAutoIncrement = "no-auto-increment"
	FlagMultisig        = "multisigaddr"
)

// GetSignCommand returns the sign command
func GetMultiSignCommand() *cobra.Command {
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
		RunE: makeMultiSignCmd(),
		Args: cobra.MinimumNArgs(3),
	}

	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().Bool(flagAmino, false, "Generate Amino encoded JSON suitable for submiting to the txs REST endpoint")
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flags.FlagChainID, "", "network chain ID")

	return cmd
}

func makeMultiSignCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		clientCtx := client.GetClientContextFromCmd(cmd)
		clientCtx, err = client.ReadTxCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}

		parsedTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
		if err != nil {
			return
		}

		txFactory := tx.NewFactoryCLI(clientCtx, cmd.Flags())
		if txFactory.SignMode() == signingtypes.SignMode_SIGN_MODE_UNSPECIFIED {
			txFactory = txFactory.WithSignMode(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
		}

		txCfg := clientCtx.TxConfig
		txBuilder, err := txCfg.WrapTxBuilder(parsedTx)
		if err != nil {
			return err
		}

		backend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)

		inBuf := bufio.NewReader(cmd.InOrStdin())
		kb, err := keyring.New(sdk.KeyringServiceName(), backend, clientCtx.HomeDir, inBuf)
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

		multisigPub := multisigInfo.GetPubKey().(*kmultisig.LegacyAminoPubKey)
		multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))
		if !clientCtx.Offline {
			accnum, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, multisigInfo.GetAddress())
			if err != nil {
				return err
			}

			txFactory = txFactory.WithAccountNumber(accnum).WithSequence(seq)
		}

		// read each signature and add it to the multisig if valid
		for i := 2; i < len(args); i++ {
			sigs, err := unmarshalSignatureJSON(clientCtx, args[i])
			if err != nil {
				return err
			}

			signingData := signing.SignerData{
				ChainID:       txFactory.ChainID(),
				AccountNumber: txFactory.AccountNumber(),
				Sequence:      txFactory.Sequence(),
			}

			for _, sig := range sigs {
				err = signing.VerifySignature(sig.PubKey, signingData, sig.Data, txCfg.SignModeHandler(), txBuilder.GetTx())
				if err != nil {
					return fmt.Errorf("couldn't verify signature: %w", err)
				}

				if err := multisig.AddSignatureV2(multisigSig, sig, multisigPub.GetPubKeys()); err != nil {
					return err
				}
			}
		}

		sigV2 := signingtypes.SignatureV2{
			PubKey:   multisigPub,
			Data:     multisigSig,
			Sequence: txFactory.Sequence(),
		}

		err = txBuilder.SetSignatures(sigV2)
		if err != nil {
			return err
		}

		sigOnly, _ := cmd.Flags().GetBool(flagSigOnly)

		aminoJSON, _ := cmd.Flags().GetBool(flagAmino)

		var json []byte

		if aminoJSON {
			stdTx, err := tx.ConvertTxToStdTx(clientCtx.LegacyAmino, txBuilder.GetTx())
			if err != nil {
				return err
			}

			req := rest.BroadcastReq{
				Tx:   stdTx,
				Mode: "block|sync|async",
			}

			json, _ = clientCtx.LegacyAmino.MarshalJSON(req)

		} else {
			json, err = marshalSignatureJSON(txCfg, txBuilder, sigOnly)
			if err != nil {
				return err
			}
		}

		outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
		if outputDoc == "" {
			cmd.Printf("%s\n", json)
			return
		}

		fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer fp.Close()

		return clientCtx.PrintString(fmt.Sprintf("%s\n", json))
	}
}

func GetBatchMultisigCommand(codec *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "multisign-batch [file] [name] [[signature-file]...]",
		Short:  "Assemble multisig transactions",
		Long:   `Assemble a batch of multisig transactions generated by batch sign command.`,
		PreRun: preSignCmd,
		RunE:   makeBatchMultisigCmd(codec),
		Args:   cobra.MinimumNArgs(3),
	}

	cmd.Flags().Bool(flagNoAutoIncrement, false, "disable sequence auto increment")
	cmd.Flags().String(
		FlagMultisig, "",
		"Address of the multisig account on behalf of which the transaction shall be signed",
	)

	return cmd
}

func makeBatchMultisigCmd(cdc *codec.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd)
		clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
		if err != nil {
			return err
		}

		parsedTxs, err := authclient.ReadTxsFromFile(clientCtx, args[0])
		if err != nil {
			return err
		}

		txFactory := tx.NewFactoryCLI(clientCtx, cmd.Flags())
		if txFactory.SignMode() == signingtypes.SignMode_SIGN_MODE_UNSPECIFIED {
			txFactory = txFactory.WithSignMode(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
		}

		txCfg := clientCtx.TxConfig
		txBldr, err := txCfg.WrapTxBuilder(parsedTxs)
		if err != nil {
			return err
		}

		inBuf := bufio.NewReader(cmd.InOrStdin())
		backend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)

		kb, err := keyring.New(sdk.KeyringServiceName(), backend, clientCtx.HomeDir, inBuf)
		if err != nil {
			return err
		}
		multisigInfo, err := kb.Key(args[1])
		if err != nil {
			return errors.Wrap(err, "error getting keybase multisig account")
		}
		if multisigInfo.GetType() != keyring.TypeMulti {
			return fmt.Errorf("%q must be of type %s: %s", args[1], crkeys.TypeMulti, multisigInfo.GetType())
		}

		multisigPub := multisigInfo.GetPubKey().(*kmultisig.LegacyAminoPubKey)
		multisigSig := multisig.NewMultisig(len(multisigPub.PubKeys))
		var signatureBatch [][]signingtypes.SignatureV2
		for i := 2; i < len(args); i++ {
			signatures, err := utils.ReadSignaturesFromFile(cdc, args[i])
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error getting signatures from file %s", args[i]))
			}

			signatureBatch = append(signatureBatch, signatures)
		}

		if !clientCtx.Offline {
			accnum, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, multisigInfo.GetAddress())
			if err != nil {
				return err
			}

			txFactory = txFactory.WithAccountNumber(accnum).WithSequence(seq)
		}
		for i, tx := range parsedTxs {
			sigBytes := types.StdSignBytes(
				txBldr.ChainID(), txBldr.AccountNumber(), txBldr.Sequence(),
				tx.Fee, tx.GetMsgs(), tx.GetMemo(),
			)

			for _, signBatch := range signatureBatch {
				if ok := signBatch[i].PubKey.VerifyBytes(sigBytes, signBatch[i].Signature); !ok {
					return fmt.Errorf("tx %d: couldn't verify signature for address %q", i, sdk.AccAddress(signBatch[i].PubKey.Address()).String())
				}
				if err := multisigSig.AddSignatureFromPubKey(signBatch[i].Signature, signBatch[i].PubKey, multisigPub.PubKeys); err != nil {
					return err
				}
			}

			newStdSig := types.StdSignature{Signature: cdc.MustMarshalBinaryBare(multisigSig), PubKey: multisigPub}
			if ok := newStdSig.VerifyBytes(sigBytes, newStdSig.Signature); !ok {
				return fmt.Errorf("error verifying multisig signature")
			}

			newTx := types.NewStdTx(tx.GetMsgs(), tx.Fee, []types.StdSignature{newStdSig}, tx.GetMemo())

			json, err := cdc.MarshalJSON(newTx)
			if err != nil {
				return errors.Wrap(err, "error marshalling tx")
			}

			_, err = fmt.Fprintf(out, "%s\n", json)
			if err != nil {
				return errors.Wrap(err, "error writing to output")
			}

			if viper.GetBool(flagNoAutoIncrement) {
				continue
			}

			sequence++
		}

		return nil
	}
}

func unmarshalSignatureJSON(clientCtx client.Context, filename string) (sigs []signingtypes.SignatureV2, err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	return clientCtx.TxConfig.UnmarshalSignatureJSON(bytes)
}
