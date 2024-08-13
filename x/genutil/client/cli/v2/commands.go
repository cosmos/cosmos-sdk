package v2

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	v2 "github.com/cosmos/cosmos-sdk/x/genutil/v2"

	"github.com/spf13/cobra"

	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// TODO(serverv2): remove app exporter notion that is server v1 specific

type genesisMM interface {
	DefaultGenesis() map[string]json.RawMessage
	ValidateGenesis(genesisData map[string]json.RawMessage) error
}

// Commands adds core sdk's sub-commands into genesis command.
func Commands(genutilModule genutil.AppModule, genMM genesisMM, appExport v2.AppExporter) *cobra.Command {
	return CommandsWithCustomMigrationMap(genutilModule, genMM, appExport, genutiltypes.MigrationMap{})
}

// CommandsWithCustomMigrationMap adds core sdk's sub-commands into genesis command with custom migration map.
// This custom migration map can be used by the application to add its own migration map.
func CommandsWithCustomMigrationMap(genutilModule genutil.AppModule, genMM genesisMM, appExport v2.AppExporter, migrationMap genutiltypes.MigrationMap) *cobra.Command {
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
		cli.CollectGenTxsCmd(genutilModule.GenTxValidator()),
		cli.ValidateGenesisCmd(genMM),
		cli.AddGenesisAccountCmd(),
		ExportCmd(appExport),
	)

	return cmd
}
