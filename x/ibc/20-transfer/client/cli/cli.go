package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// GetQueryCmd returns the query commands for IBC fungible token transfer
func GetQueryCmd(cdc *codec.Codec, queryRoute string) *cobra.Command {
	ics20TransferQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "IBC fungible token transfer query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics20TransferQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryNextSequence(cdc, queryRoute),
	)...)

	return ics20TransferQueryCmd
}

// NewTxCmd returns a root CLI command handler for all x/ibc fungible token transfer
func NewTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	ics20TransferTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "IBC fungible token transfer transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics20TransferTxCmd.AddCommand(flags.PostCommands(
		NewTransferTxCmd(m, txg, ar),
	)...)

	return ics20TransferTxCmd
}
