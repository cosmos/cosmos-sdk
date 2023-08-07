package cli

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const flagGenesisTime = "genesis-time"

// MigrationMap is a map of SDK versions to their respective genesis migration functions.
var MigrationMap = types.MigrationMap{}

// MigrateGenesisCmd returns a command to execute genesis state migration.
// Applications should pass their own migration map to this function.
// When the application migration includes a SDK migration, the Cosmos SDK migration function should as well be called.
func MigrateGenesisCmd(migrations types.MigrationMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate <target-version> <genesis-file>",
		Short:   "Migrate genesis to a specified target version",
		Long:    "Migrate the source genesis into the target version and print to STDOUT",
		Example: fmt.Sprintf("%s migrate v0.47 /path/to/genesis.json --chain-id=cosmoshub-3 --genesis-time=2019-04-22T17:00:00Z", version.AppName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return MigrateHandler(cmd, args, migrations)
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "Override genesis_time with this flag")
	cmd.Flags().String(flags.FlagChainID, "", "Override chain_id with this flag")
	cmd.Flags().String(flags.FlagOutputDocument, "", "Exported state is written to the given file instead of STDOUT")

	return cmd
}

// MigrateHandler handles the migration command with a migration map as input,
// returning an error upon failure.
func MigrateHandler(cmd *cobra.Command, args []string, migrations types.MigrationMap) error {
	clientCtx := client.GetClientContextFromCmd(cmd)

	target := args[0]
	migrationFunc, ok := migrations[target]
	if !ok || migrationFunc == nil {
		versions := maps.Keys(migrations)
		return fmt.Errorf("unknown migration function for version: %s (supported versions %s)", target, strings.Join(slices.Sorted(versions), ", "))
	}

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
			return MigrateHandler(cmd, args, migrationMap)
		},
	}

	newGenState, err := migrationFunc(initialState, clientCtx)
	if err != nil {
		return fmt.Errorf("failed to migrate genesis state: %w", err)
	}

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
}

// MigrateHandler handles the migration command with a migration map as input,
// returning an error upon failure.
func MigrateHandler(cmd *cobra.Command, args []string, migrations types.MigrationMap) error {
	clientCtx := client.GetClientContextFromCmd(cmd)

	var err error

	target := args[0]
	importGenesis := args[1]

	genDoc, err := validateGenDoc(importGenesis)
	if err != nil {
		return err
	}

	// Since some default values are valid values, we just print to
	// make sure the user didn't forget to update these values.
	if genDoc.ConsensusParams.Evidence.MaxBytes == 0 {
		fmt.Printf("Warning: consensus_params.evidence.max_bytes is set to 0. If this is"+
			" deliberate, feel free to ignore this warning. If not, please have a look at the chain"+
			" upgrade guide at %s.\n", chainUpgradeGuide)
	}

	var initialState types.AppMap
	if err := json.Unmarshal(genDoc.AppState, &initialState); err != nil {
		return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
	}

	migrationFunc := migrations[target]
	if migrationFunc == nil {
		return fmt.Errorf("unknown migration function for version: %s", target)
	}

	// TODO: handler error from migrationFunc call
	newGenState := migrationFunc(initialState, clientCtx)

	genDoc.AppState, err = json.Marshal(newGenState)
	if err != nil {
		return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
	}

	genesisTime, _ := cmd.Flags().GetString(flagGenesisTime)
	if genesisTime != "" {
		var t time.Time

		err := t.UnmarshalText([]byte(genesisTime))
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal genesis time")
		}

		genDoc.GenesisTime = t
	}

	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	if chainID != "" {
		genDoc.ChainID = chainID
	}

	bz, err := tmjson.Marshal(genDoc)
	if err != nil {
		return errors.Wrap(err, "failed to marshal genesis doc")
	}

	sortedBz, err := sdk.SortJSON(bz)
	if err != nil {
		return errors.Wrap(err, "failed to sort JSON genesis doc")
	}

	cmd.Println(string(sortedBz))
	return nil
}
