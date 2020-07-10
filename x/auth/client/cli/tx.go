package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(clientCtx client.Context) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetMultiSignCommand(clientCtx),
		GetSignCommand(clientCtx),
		GetValidateSignaturesCommand(clientCtx),
		GetSignBatchCommand(),
	)
	return txCmd
}
