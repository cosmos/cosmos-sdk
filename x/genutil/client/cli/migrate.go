package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	v043 "github.com/cosmos/cosmos-sdk/x/genutil/migrations/v043"
	v046 "github.com/cosmos/cosmos-sdk/x/genutil/migrations/v046"
	v047 "github.com/cosmos/cosmos-sdk/x/genutil/migrations/v047"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const flagGenesisTime = "genesis-time"

// Allow applications to extend and modify the migration process.
//
// Ref: https://github.com/cosmos/cosmos-sdk/issues/5041
var migrationMap = types.MigrationMap{
	"v0.43": v043.Migrate, // NOTE: v0.43, v0.44 and v0.45 are genesis compatible.
	"v0.46": v046.Migrate,
	"v0.47": v047.Migrate,
}

// GetMigrationCallback returns a MigrationCallback for a given version.
func GetMigrationCallback(version string) types.MigrationCallback {
	return migrationMap[version]
}

// GetMigrationVersions get all migration version in a sorted slice.
func GetMigrationVersions() []string {
	versions := maps.Keys(migrationMap)
	sort.Strings(versions)

	return versions
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] [genesis-file]",
		Short: "Migrate genesis to a specified target version",
		Long: fmt.Sprintf(`Migrate the source genesis into the target version and print to STDOUT.

Example:
$ %s migrate v0.36 /path/to/genesis.json --chain-id=cosmoshub-3 --genesis-time=2019-04-22T17:00:00Z
`, version.AppName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			var err error

			target := args[0]
			importGenesis := args[1]

			appGenesis, err := types.AppGenesisFromFile(importGenesis)
			if err != nil {
				return err
			}

			if err := appGenesis.ValidateAndComplete(); err != nil {
				return fmt.Errorf("make sure that you have correctly migrated all CometBFT consensus params. Refer the UPGRADING.md (%s): %w", chainUpgradeGuide, err)
			}

			// Since some default values are valid values, we just print to
			// make sure the user didn't forget to update these values.
			if appGenesis.Consensus.Params.Evidence.MaxBytes == 0 {
				fmt.Printf("Warning: consensus.params.evidence.max_bytes is set to 0. If this is"+
					" deliberate, feel free to ignore this warning. If not, please have a look at the chain"+
					" upgrade guide at %s.\n", chainUpgradeGuide)
			}

			var initialState types.AppMap
			if err := json.Unmarshal(appGenesis.AppState, &initialState); err != nil {
				return fmt.Errorf("failed to JSON unmarshal initial genesis state: %w", err)
			}

			migrationFunc := GetMigrationCallback(target)
			if migrationFunc == nil {
				return fmt.Errorf("unknown migration function for version: %s", target)
			}

			// TODO: handler error from migrationFunc call
			newGenState := migrationFunc(initialState, clientCtx)

			appGenesis.AppState, err = json.Marshal(newGenState)
			if err != nil {
				return fmt.Errorf("failed to JSON marshal migrated genesis state: %w", err)
			}

			genesisTime, _ := cmd.Flags().GetString(flagGenesisTime)
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return fmt.Errorf("failed to unmarshal genesis time: %w", err)
				}

				appGenesis.GenesisTime = t
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID != "" {
				appGenesis.ChainID = chainID
			}

			bz, err := json.Marshal(appGenesis)
			if err != nil {
				return fmt.Errorf("failed to marshal app genesis: %w", err)
			}

			outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			if outputDocument == "" {
				cmd.Println(string(bz))
				return nil
			}

			if err = appGenesis.SaveAs(outputDocument); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "Override genesis_time with this flag")
	cmd.Flags().String(flags.FlagChainID, "", "Override chain_id with this flag")
	cmd.Flags().String(flags.FlagOutputDocument, "", "Exported state is written to the given file instead of STDOUT")

	return cmd
}
