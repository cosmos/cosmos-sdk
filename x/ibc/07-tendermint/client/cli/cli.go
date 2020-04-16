package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// NewTxCmd returns a root CLI command handler for all x/ibc tendermint client transaction commands.
func NewTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	ics07TendermintTxCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC tendermint transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics07TendermintTxCmd.AddCommand(flags.PostCommands(
		NewCreateClientTxCmd(m, txg, ar),
		NewUpdateClientTxCmd(m, txg, ar),
		NewSubmitMisbehaviourTxCmd(m, txg, ar),
	)...)

	return ics07TendermintTxCmd
}
