package offchain

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

const (
	flagNotEmitUnpopulated = "notEmitUnpopulated"
	flagIndent             = "indent"
	flagEncoding           = "encoding"
)

// OffChain off-chain utilities.
func OffChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offchain",
		Short: "Off-chain utilities.",
		Long:  `Utilities for off-chain data.`,
	}

	cmd.AddCommand(
		SignFile(),
		VerifyFile(),
	)

	flags.AddKeyringFlags(cmd.PersistentFlags())
	return cmd
}

// SignFile signs a file with a key.
func SignFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-file <keyName> <fileName>",
		Short: "Sign a file.",
		Long:  "Sign a file using a given key.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			signmode, _ := cmd.Flags().GetString(flags.FlagSignMode)
			notEmitUnpopulated, _ := cmd.Flags().GetBool(flagNotEmitUnpopulated)
			indent, _ := cmd.Flags().GetString(flagIndent)
			encoding, _ := cmd.Flags().GetString(flagEncoding)

			signedTx, err := Sign(clientCtx, bz, args[0], signmode, indent, encoding, !notEmitUnpopulated)
			if err != nil {
				return err
			}

			cmd.Println(signedTx)
			return nil
		},
	}

	cmd.PersistentFlags().String(flagIndent, "  ", "Choose an indent for the tx")
	cmd.PersistentFlags().Bool(flagNotEmitUnpopulated, false, "Don't show unpopulated fields in the tx")
	cmd.PersistentFlags().String(flags.FlagSignMode, "direct", "Choose sign mode (direct|amino-json|direct-aux|textual), this is an advanced feature")
	cmd.PersistentFlags().String(flagEncoding, "no-encoding", "Choose an encoding method for the file content to be added as the tx data (no-encoding|base64)")
	return cmd
}

// VerifyFile verifies given file with given key.
func VerifyFile() *cobra.Command {
	return &cobra.Command{
		Use:   "verify-file <keyName> <fileName>",
		Short: "Verify a file.",
		Long:  "Verify a previously signed file with the given key.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			err = Verify(clientCtx, bz)
			if err == nil {
				cmd.Println("Verification OK!")
			}
			return err
		},
	}
}
