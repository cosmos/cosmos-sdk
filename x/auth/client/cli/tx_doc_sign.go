package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

func GetSignDocCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-doc [file]",
		Short: "Sign data in the form of StdSignDoc",
		Long:  `Sign data in the form of StdSignDoc https://github.com/cosmos/cosmos-sdk/blob/f7c631eef9361165cfd8eec98fb783858acfa0d7/x/auth/types/stdtx.go#L216-L223`,
		RunE:  makeSignDocCmd(),
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")
	cmd.Flags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func makeSignDocCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		doc, err := readStdSignDocFromFile(args[0])
		if err != nil {
			return err
		}

		keybase := tx.NewFactoryCLI(ctx, cmd.Flags()).Keybase()
		sig, err := signStdSignDoc(ctx, keybase, doc)
		if err != nil {
			return err
		}

		json := legacy.Cdc.MustMarshalJSON(sig)

		if viper.GetString(flags.FlagOutputDocument) == "" {
			fmt.Printf("%s\n", json)
			return nil
		}

		fp, err := os.OpenFile(
			viper.GetString(flags.FlagOutputDocument), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644,
		)
		if err != nil {
			return err
		}

		defer fp.Close()
		fmt.Fprintf(fp, "%s\n", json)

		return nil
	}
}

// Read StdSignDoc from the given filename.  Can pass "-" to read from stdin.
func readStdSignDocFromFile(filename string) (doc legacytx.StdSignDoc, err error) {
	var bytes []byte

	if filename == "-" {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(filename)
	}

	if err != nil {
		return
	}

	legacy.Cdc.MustUnmarshalJSON(bytes, &doc)

	return
}

// SignStdTxWithSignerAddress attaches a signature to a StdTx and returns a copy of a it.
// Don't perform online validation or lookups if offline is true, else
// populate account and sequence numbers from a foreign account.
func signStdSignDoc(ctx client.Context, keybase keyring.Keyring, doc legacytx.StdSignDoc) (sig legacytx.StdSignature, err error) { //nolint:staticcheck // this will be removed when proto is ready

	sig, err = makeSignature(keybase, ctx.GetFromName(), doc)
	if err != nil {
		return legacytx.StdSignature{}, err //nolint:staticcheck // this will be removed when proto is ready
	}

	return sig, nil
}

func makeSignature(keybase keyring.Keyring, name string, doc legacytx.StdSignDoc) (legacytx.StdSignature, error) { //nolint:staticcheck // this will be removed when proto is ready
	bz := sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(doc))

	sigBytes, pubkey, err := keybase.Sign(name, bz)
	if err != nil {
		return legacytx.StdSignature{}, err //nolint:staticcheck // this will be removed when proto is ready
	}
	return legacytx.StdSignature{ //nolint:staticcheck // this will be removed when proto is ready
		PubKey:    pubkey,
		Signature: sigBytes,
	}, nil
}
