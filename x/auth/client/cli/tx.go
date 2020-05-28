package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(ctx context.CLIContext) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetMultiSignCommand(ctx),
		GetSignCommand(ctx),
		GetValidateSignaturesCommand(ctx),
	)
	return txCmd
}
