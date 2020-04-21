package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	tmclient "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/client/cli"
	localhostclient "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "IBC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ibcTxCmd.AddCommand(flags.PostCommands(
		tmclient.GetTxCmd(cdc, storeKey),
		localhostclient.GetTxCmd(cdc, storeKey),
		connection.GetTxCmd(cdc, storeKey),
		channel.GetTxCmd(cdc, storeKey),
	)...)
	return ibcTxCmd
}

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group ibc queries under a subcommand
	ibcQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the IBC module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ibcQueryCmd.AddCommand(flags.GetCommands(
		ibcclient.GetQueryCmd(cdc, queryRoute),
		connection.GetQueryCmd(cdc, queryRoute),
		channel.GetQueryCmd(cdc, queryRoute),
	)...)
	return ibcQueryCmd
}
