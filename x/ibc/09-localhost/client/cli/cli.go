package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// GetTxCmd returns the transaction commands for IBC
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	ics09LocalhostTxCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "Localhost transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics09LocalhostTxCmd.AddCommand(flags.PostCommands(
		GetCmdCreateClient(cdc),
	)...)

	return ics09LocalhostTxCmd
}
