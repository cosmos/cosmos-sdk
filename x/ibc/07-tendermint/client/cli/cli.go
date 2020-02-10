package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for IBC clients
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	ics07TendermintTxCmd := &cobra.Command{
		Use:                        "tendermint",
		Short:                      "Tendermint transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics07TendermintTxCmd.AddCommand(flags.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
		GetCmdSubmitMisbehaviour(cdc),
	)...)

	return ics07TendermintTxCmd
}
