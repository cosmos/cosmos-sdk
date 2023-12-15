package offchain

import (
	"os"

	"github.com/spf13/cobra"

	v2flags "cosmossdk.io/client/v2/internal/flags"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

const (
	flagNotEmitUnpopulated = "notEmitUnpopulated"
	flagIndent             = "indent"
	flagEncoding           = "encoding"
	flagFileFormat         = "file-format"
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
			outputFormat, _ := cmd.Flags().GetString(v2flags.FlagOutput)

			signedTx, err := Sign(clientCtx, bz, args[0], signmode, indent, encoding, outputFormat, !notEmitUnpopulated)
			if err != nil {
				return err
			}

			cmd.Println(signedTx)
			return nil
		},
	}

	cmd.PersistentFlags().String(flagIndent, "  ", "Choose an indent for the tx")
	cmd.PersistentFlags().String(v2flags.FlagOutput, "json", "Choose output format (json|text")
	cmd.PersistentFlags().Bool(flagNotEmitUnpopulated, false, "Don't show unpopulated fields in the tx")
	cmd.PersistentFlags().String(flags.FlagSignMode, "direct", "Choose sign mode (direct|amino-json|direct-aux|textual), this is an advanced feature")
	cmd.PersistentFlags().String(flagEncoding, "no-encoding", "Choose an encoding method for the file content to be added as the tx data (no-encoding|base64)")
	return cmd
}

// VerifyFile verifies given file with given key.
func VerifyFile() *cobra.Command {
	cmd := &cobra.Command{
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

			fileFormat, _ := cmd.Flags().GetString(flagFileFormat)

			err = Verify(clientCtx, bz, fileFormat)
			if err == nil {
				cmd.Println("Verification OK!")
			}
			return err
		},
	}

	cmd.PersistentFlags().String(flagFileFormat, "json", "Choose whats the file format to be verified (json|text)")
	return cmd
}
