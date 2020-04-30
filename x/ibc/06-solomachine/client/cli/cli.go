package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for IBC clients
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	ics06SoloMachineTxCmd := &cobra.Command{
		Use:                        "solomachine",
		Short:                      "Solo Machine transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics06SoloMachineTxCmd.AddCommand(flags.PostCommand(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
		GetCmdSubmitMisbehaviour(cdc),
	)...)

	return ics06SoloMachineTxCmd
}
