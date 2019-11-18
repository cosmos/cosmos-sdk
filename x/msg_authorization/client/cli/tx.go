package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	AuthorizationTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Authorization transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	AuthorizationTxCmd.AddCommand(client.PostCommands(
		GetCmdGrantCapability(cdc),
		GetCmdRevokeCapability(cdc),
	)...)

	return AuthorizationTxCmd
}

func GetCmdGrantCapability(cdc *codec.Codec) *cobra.Command {
	//TODO
	return nil
}

func GetCmdRevokeCapability(cdc *codec.Codec) *cobra.Command {
	//TODO
	return nil
}
