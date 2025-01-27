package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/transaction"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"
)

type genesisMM interface {
	DefaultGenesis() map[string]json.RawMessage
	ValidateGenesis(genesisData map[string]json.RawMessage) error
}

type ExportableApp interface {
	ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (v2.ExportedApp, error)
	LoadHeight(uint64) error
}

// Commands adds core sdk's sub-commands into genesis command.
func Commands(
	genTxValidator func([]transaction.Msg) error,
	genMM genesisMM,
	exportable ExportableApp,
) *cobra.Command {
	return CommandsWithCustomMigrationMap(genTxValidator, genMM, exportable, cli.MigrationMap)
}

// CommandsWithCustomMigrationMap adds core sdk's sub-commands into genesis command with custom migration map.
// This custom migration map can be used by the application to add its own migration map.
func CommandsWithCustomMigrationMap(
	genTxValidator func([]transaction.Msg) error,
	genMM genesisMM,
	exportable ExportableApp,
	migrationMap genutiltypes.MigrationMap,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "genesis",
		Short:                      "Application's genesis-related subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		cli.GenTxCmd(genMM, banktypes.GenesisBalancesIterator{}),
		cli.MigrateGenesisCmd(migrationMap),
		cli.CollectGenTxsCmd(genTxValidator),
		cli.ValidateGenesisCmd(genMM),
		cli.AddGenesisAccountCmd(),
		ExportCmd(exportable),
	)

	return cmd
}
