package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// TODO(serverv2): remove app exporter notion that is server v1 specific

type genesisMM interface {
	DefaultGenesis() map[string]json.RawMessage
	ValidateGenesis(genesisData map[string]json.RawMessage) error
}

// Commands adds core sdk's sub-commands into genesis command.
func Commands(txConfig client.TxConfig, genutilModule genutil.AppModule, genMM genesisMM, appExport servertypes.AppExporter) *cobra.Command {
	return CommandsWithCustomMigrationMap(txConfig, genutilModule, genMM, appExport, MigrationMap)
}

// CommandsWithCustomMigrationMap adds core sdk's sub-commands into genesis command with custom migration map.
// This custom migration map can be used by the application to add its own migration map.
func CommandsWithCustomMigrationMap(txConfig client.TxConfig, genutilModule genutil.AppModule, genMM genesisMM, appExport servertypes.AppExporter, migrationMap genutiltypes.MigrationMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "genesis",
		Short:                      "Application's genesis-related subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		GenTxCmd(genMM, txConfig, banktypes.GenesisBalancesIterator{}, txConfig.SigningContext().ValidatorAddressCodec()),
		MigrateGenesisCmd(migrationMap),
		CollectGenTxsCmd(genutilModule.GenTxValidator()),
		ValidateGenesisCmd(genMM),
		AddGenesisAccountCmd(txConfig.SigningContext().AddressCodec()),
		ExportCmd(appExport),
	)

	return cmd
}
