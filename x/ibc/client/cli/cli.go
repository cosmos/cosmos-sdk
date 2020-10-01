package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	solomachine "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        host.ModuleName,
		Short:                      "IBC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ibcTxCmd.AddCommand(
		solomachine.GetTxCmd(),
		tendermint.GetTxCmd(),
		connection.GetTxCmd(),
		channel.GetTxCmd(),
	)

	return ibcTxCmd
}

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group ibc queries under a subcommand
	ibcQueryCmd := &cobra.Command{
		Use:                        host.ModuleName,
		Short:                      "Querying commands for the IBC module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ibcQueryCmd.AddCommand(
		ibcclient.GetQueryCmd(),
		connection.GetQueryCmd(),
		channel.GetQueryCmd(),
	)

	return ibcQueryCmd
}
