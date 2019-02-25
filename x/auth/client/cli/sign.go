package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

const (
	flagMultisig     = "multisig"
	flagAppend       = "append"
	flagValidateSigs = "validate-signatures"
	flagOffline      = "offline"
	flagSigOnly      = "signature-only"
	flagOutfile      = "output-document"
)

// GetSignCommand returns the sign command
func GetSignCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [file]",
		Short: "Sign transactions generated offline",
		Long: `Sign transactions created with the --generate-only flag.
Read a transaction from [file], sign it, and print its JSON encoding.

If the flag --signature-only flag is on, it outputs a JSON representation
of the generated signature only.

If the flag --validate-signatures is on, then the command would check whether all required
signers have signed the transactions, whether the signatures were collected in the right
order, and if the signature is valid over the given transaction. If the --offline
flag is also provided, signature validation over the transaction will be not be
performed as that will require communication with a full node.

The --offline flag makes sure that the client will not reach out to an external node.
Thus account number or sequence number lookups will not be performed and it is
recommended to set such parameters manually.

The --multisig=<multisig_key> flag generates a signature on behalf of a multisig account
key. It implies --signature-only. Full multisig signed transactions may eventually
be generated via the 'multisign' command.
`,
		RunE: makeSignCmd(codec),
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String(client.FlagName, "", "Name of private key with which to sign")
	cmd.Flags().String(flagMultisig, "",
		"Address of the multisig account on behalf of which the "+
			"transaction shall be signed")
	cmd.Flags().Bool(flagAppend, true,
		"Append the signature to the existing ones. "+
			"If disabled, old signatures would be overwritten. Ignored if --multisig is on")
	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit")
	cmd.Flags().Bool(flagValidateSigs, false, "Print the addresses that must sign the transaction, "+
		"those who have already signed it, and make sure that signatures are in the correct order")
	cmd.Flags().Bool(flagOffline, false, "Offline mode. Do not query a full node")
	cmd.Flags().String(flagOutfile, "",
		"The document will be written to the given file instead of STDOUT")

	// Add the flags here and return the command
	return client.PostCommands(cmd)[0]
}

func makeSignCmd(cdc *amino.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		stdTx, err := authclient.ReadStdTxFromFile(cdc, args[0])
		if err != nil {
			return
		}

		offline := viper.GetBool(flagOffline)
		cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
		txBldr := authtxb.NewTxBuilderFromCLI()

		if viper.GetBool(flagValidateSigs) {
			if !printAndValidateSigs(cliCtx, txBldr.ChainID(), stdTx, offline) {
				return fmt.Errorf("signatures validation failed")
			}

			return nil
		}

		name := viper.GetString(client.FlagName)
		if name == "" {
			return errors.New("required flag \"name\" has not been set")
		}

		// if --signature-only is on, then override --append
		var newTx auth.StdTx
		generateSignatureOnly := viper.GetBool(flagSigOnly)
		multisigAddrStr := viper.GetString(flagMultisig)

		if multisigAddrStr != "" {
			var multisigAddr sdk.AccAddress
			multisigAddr, err = sdk.AccAddressFromBech32(multisigAddrStr)
			if err != nil {
				return err
			}

			newTx, err = utils.SignStdTxWithSignerAddress(
				txBldr, cliCtx, multisigAddr, name, stdTx, offline)
			generateSignatureOnly = true
		} else {
			appendSig := viper.GetBool(flagAppend) && !generateSignatureOnly
			newTx, err = utils.SignStdTx(
				txBldr, cliCtx, name, stdTx, appendSig, offline)
		}
		if err != nil {
			return err
		}

		var json []byte

		switch generateSignatureOnly {
		case true:
			switch cliCtx.Indent {
			case true:
				json, err = cdc.MarshalJSONIndent(newTx.Signatures[0], "", "  ")
			default:
				json, err = cdc.MarshalJSON(newTx.Signatures[0])
			}
		default:
			switch cliCtx.Indent {
			case true:
				json, err = cdc.MarshalJSONIndent(newTx, "", "  ")
			default:
				json, err = cdc.MarshalJSON(newTx)
			}
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

// printAndValidateSigs will validate the signatures of a given transaction over
// its expected signers. In addition, if offline has not been supplied, the
// signature is verified over the transaction sign bytes.
func printAndValidateSigs(
	cliCtx context.CLIContext, chainID string, stdTx auth.StdTx, offline bool,
) bool {

	fmt.Println("Signers:")

	signers := stdTx.GetSigners()
	for i, signer := range signers {
		fmt.Printf(" %v: %v\n", i, signer.String())
	}

	success := true
	sigs := stdTx.GetSignatures()

	fmt.Println("")
	fmt.Println("Signatures:")

	if len(sigs) != len(signers) {
		success = false
	}

	for i, sig := range sigs {
		sigAddr := sdk.AccAddress(sig.Address())
		sigSanity := "OK"

		if i >= len(signers) || !sigAddr.Equals(signers[i]) {
			sigSanity = "ERROR: signature does not match its respective signer"
			success = false
		}

		// Validate the actual signature over the transaction bytes since we can
		// reach out to a full node to query accounts.
		if !offline && success {
			acc, err := cliCtx.GetAccount(sigAddr)
			if err != nil {
				fmt.Printf("failed to get account: %s\n", sigAddr)
				return false
			}

			sigBytes := auth.StdSignBytes(
				chainID, acc.GetAccountNumber(), acc.GetSequence(),
				stdTx.Fee, stdTx.GetMsgs(), stdTx.GetMemo(),
			)

			if ok := sig.VerifyBytes(sigBytes, sig.Signature); !ok {
				sigSanity = "ERROR: signature invalid"
				success = false
			}
		}

		fmt.Printf(" %v: %v\t[%s]\n", i, sigAddr.String(), sigSanity)
	}

	fmt.Println("")
	return success
}
