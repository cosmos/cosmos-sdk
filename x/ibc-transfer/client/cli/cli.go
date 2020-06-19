package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the query commands for IBC fungible token transfer
func GetQueryCmd(clientCtx client.Context) *cobra.Command {
	ics20TransferQueryCmd := &cobra.Command{
		Use:                        "ibc-transfer",
		Short:                      "IBC fungible token transfer query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics20TransferQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryNextSequence(clientCtx),
	)...)

	return ics20TransferQueryCmd
}

// NewTxCmd returns the transaction commands for IBC fungible token transfer
func NewTxCmd(clientCtx client.Context) *cobra.Command {
	ics20TransferTxCmd := &cobra.Command{
		Use:                        "ibc-transfer",
		Short:                      "IBC fungible token transfer transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics20TransferTxCmd.AddCommand(flags.PostCommands(
		NewTransferTxCmd(clientCtx),
	)...)

	return ics20TransferTxCmd
}
