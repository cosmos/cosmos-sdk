package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
)

const (
	flagAppend       = "append"
	flagValidateSigs = "validate-signatures"
	flagOffline      = "offline"
	flagSigOnly      = "signature-only"
	flagOutfile      = "output-document"
)

// GetSignCommand returns the sign command
func GetSignCommand(codec *amino.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign <file>",
		Short: "Sign transactions generated offline",
		Long: `Sign transactions created with the --generate-only flag.
Read a transaction from <file>, sign it, and print its JSON encoding.

If the flag --signature-only flag is on, it outputs a JSON representation
of the generated signature only.

If the flag --validate-signatures is on, then the command would check whether all required
signers have signed the transactions and whether the signatures were collected in the right
order.

The --offline flag makes sure that the client will not reach out to the local cache.
Thus account number or sequence number lookups will not be performed and it is
recommended to set such parameters manually.`,
		RunE: makeSignCmd(codec),
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String(client.FlagName, "", "Name of private key with which to sign")
	cmd.Flags().Bool(flagAppend, true,
		"Append the signature to the existing ones. If disabled, old signatures would be overwritten")
	cmd.Flags().Bool(flagSigOnly, false, "Print only the generated signature, then exit.")
	cmd.Flags().Bool(flagValidateSigs, false, "Print the addresses that must sign the transaction, "+
		"those who have already signed it, and make sure that signatures are in the correct order.")
	cmd.Flags().Bool(flagOffline, false, "Offline mode. Do not query local cache.")
	cmd.Flags().String(flagOutfile, "",
		"The document will be written to the given file instead of STDOUT")

	// Add the flags here and return the command
	return client.PostCommands(cmd)[0]
}

func makeSignCmd(cdc *amino.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		stdTx, err := readAndUnmarshalStdTx(cdc, args[0])
		if err != nil {
			return
		}

		if viper.GetBool(flagValidateSigs) {
			if !printSignatures(stdTx) {
				return fmt.Errorf("signatures validation failed")
			}
			return nil
		}

		name := viper.GetString(client.FlagName)
		if name == "" {
			return errors.New("required flag \"name\" has not been set")
		}
		cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
		txBldr := authtxb.NewTxBuilderFromCLI()

		// if --signature-only is on, then override --append
		generateSignatureOnly := viper.GetBool(flagSigOnly)
		appendSig := viper.GetBool(flagAppend) && !generateSignatureOnly
		newTx, err := utils.SignStdTx(txBldr, cliCtx, name, stdTx, appendSig, viper.GetBool(flagOffline))
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

func printSignatures(stdTx auth.StdTx) bool {
	fmt.Println("Signers:")
	signers := stdTx.GetSigners()
	for i, signer := range signers {
		fmt.Printf(" %v: %v\n", i, signer.String())
	}

	sigs := stdTx.GetSignatures()
	fmt.Println("")
	fmt.Println("Signatures:")
	success := true
	if len(sigs) != len(signers) {
		success = false
	}
	for i, sig := range stdTx.GetSignatures() {
		sigAddr := sdk.AccAddress(sig.Address())
		sigSanity := "OK"
		if i >= len(signers) || !sigAddr.Equals(signers[i]) {
			sigSanity = fmt.Sprintf("ERROR: signature %d does not match its respective signer", i)
			success = false
		}
		fmt.Printf(" %v: %v\t[%s]\n", i, sigAddr.String(), sigSanity)
	}
	fmt.Println("")
	return success
}

func readAndUnmarshalStdTx(cdc *amino.Codec, filename string) (stdTx auth.StdTx, err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if err = cdc.UnmarshalJSON(bytes, &stdTx); err != nil {
		return
	}
	return
}
