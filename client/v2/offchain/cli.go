package offchain

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// TODO: auto-cli

// OffChain off-chain utilities
func OffChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offchain",
		Short: "offchain msg utilities",
		Long:  ``,
	}

	cmd.AddCommand(
		SignFile(),
		VerifyFile(),
	)

	cmd.PersistentFlags().String(flags.FlagOutput, "text", "Output format (text|json)")
	flags.AddKeyringFlags(cmd.PersistentFlags())

	return cmd
}

// SignFile sign a file with a key
func SignFile() *cobra.Command {
	return &cobra.Command{
		Use:   "sign <keyName> <fileName>",
		Short: "Sign a file",
		Long:  "Sign a file using a key.",
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

			signedTx, err := Sign(clientCtx, bz, args[0], clientCtx.SignModeStr)
			if err != nil {
				return err
			}
			cmd.Println(signedTx)
			return nil
		},
	}
}

// VerifyFile sign a file with a key
func VerifyFile() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <keyName> <fileName>",
		Short: "",
		Long:  "",
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

			return Verify(clientCtx, bz)
		},
	}
}
