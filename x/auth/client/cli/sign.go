package cli

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

const (
	flagAppend    = "append"
	flagPrintSigs = "print-sigs"
)

// GetSignCommand returns the sign command
func GetSignCommand(codec *amino.Codec, decoder auth.AccountDecoder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign <file>",
		Short: "Sign transactions",
		Long: `Sign transactions created with the --generate-only flag.
Read a transaction from <file>, sign it, and print its JSON encoding.`,
		RunE: makeSignCmd(codec, decoder),
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String(client.FlagName, "", "Name of private key with which to sign")
	cmd.Flags().Bool(flagAppend, true, "Append the signature to the existing ones. If disabled, old signatures would be overwritten")
	cmd.Flags().Bool(flagPrintSigs, false, "Print the addresses that must sign the transaction and those who have already signed it, then exit")
	return cmd
}

func makeSignCmd(cdc *amino.Codec, decoder auth.AccountDecoder) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		stdTx, err := readAndUnmarshalStdTx(cdc, args[0])
		if err != nil {
			return
		}

		if viper.GetBool(flagPrintSigs) {
			printSignatures(stdTx)
			return nil
		}

		name := viper.GetString(client.FlagName)
		cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(decoder)
		txBldr := authtxb.NewTxBuilderFromCLI()

		newTx, err := utils.SignStdTx(txBldr, cliCtx, name, stdTx, viper.GetBool(flagAppend))
		if err != nil {
			return err
		}
		json, err := cdc.MarshalJSON(newTx)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", json)
		return
	}
}

func printSignatures(stdTx auth.StdTx) {
	fmt.Println("Signers:")
	for i, signer := range stdTx.GetSigners() {
		fmt.Printf(" %v: %v\n", i, signer.String())
	}
	fmt.Println("")
	fmt.Println("Signatures:")
	for i, sig := range stdTx.GetSignatures() {
		fmt.Printf(" %v: %v\n", i, sdk.AccAddress(sig.Address()).String())
	}
	return
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
