package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ModuleHashByHeightQuery retrieves the module hashes at a given height.
func ModuleHashByHeightQuery[T servertypes.Application](appCreator servertypes.AppCreator[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-hash-by-height [height]",
		Short: "Get module hashes at a given height",
		Long: `Get module hashes at a given height. This command is useful for debugging and verifying the state of the application at a given height. Daemon should not be running when calling this command.
Example:
	osmosisd module-hash-by-height 16841115,
`,
		Args: cobra.ExactArgs(1), // Ensure exactly one argument is provided
		RunE: func(cmd *cobra.Command, args []string) error {
			heightToRetrieveString := args[0]

			serverCtx := GetServerContextFromCmd(cmd)

			height, err := strconv.ParseInt(heightToRetrieveString, 10, 64)
			if err != nil {
				return err
			}

			commitInfoForHeight, err := getModuleHashesAtHeight(serverCtx, appCreator, height)
			if err != nil {
				return err
			}

			// Get the output format flag
			outputFormat, err := cmd.Flags().GetString(flags.FlagOutput)
			if err != nil {
				return err
			}

			// Print the CommitInfo based on the output format
			switch outputFormat {
			case flags.OutputFormatJSON:
				jsonOutput, err := json.MarshalIndent(commitInfoForHeight, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(jsonOutput))
			case flags.OutputFormatText:
				fallthrough
			default:
				fmt.Println(commitInfoForHeight.String())
			}

			return nil
		},
	}

	cmd.Flags().StringP(flags.FlagOutput, "o", flags.OutputFormatText, "Output format (text|json)")

	return cmd
}

func getModuleHashesAtHeight[T servertypes.Application](svrCtx *Context, appCreator servertypes.AppCreator[T], height int64) (*storetypes.CommitInfo, error) {
	home := svrCtx.Config.RootDir
	db, err := openDB(home, GetAppDBBackend(svrCtx.Viper))
	if err != nil {
		return nil, fmt.Errorf("error opening DB, make sure osmosisd is not running when calling this query: %w", err)
	}
	app := appCreator(svrCtx.Logger, db, nil, svrCtx.Viper)
	rms, ok := app.CommitMultiStore().(*rootmulti.Store)
	if !ok {
		return nil, fmt.Errorf("expected rootmulti.Store, got %T", app.CommitMultiStore())
	}

	commitInfoForHeight, err := rms.GetCommitInfo(height)
	if err != nil {
		return nil, err
	}

	// Create a new slice of StoreInfos for storing the modified hashes.
	storeInfos := make([]storetypes.StoreInfo, len(commitInfoForHeight.StoreInfos))

	for i, storeInfo := range commitInfoForHeight.StoreInfos {
		// Convert the hash to a hexadecimal string.
		hash := strings.ToUpper(hex.EncodeToString(storeInfo.CommitId.Hash))

		// Create a new StoreInfo with the modified hash.
		storeInfos[i] = storetypes.StoreInfo{
			Name: storeInfo.Name,
			CommitId: storetypes.CommitID{
				Version: storeInfo.CommitId.Version,
				Hash:    []byte(hash),
			},
		}
	}

	// Sort the storeInfos slice based on the module name.
	sort.Slice(storeInfos, func(i, j int) bool {
		return storeInfos[i].Name < storeInfos[j].Name
	})

	// Create a new CommitInfo with the modified StoreInfos.
	commitInfoForHeight = &storetypes.CommitInfo{
		Version:    commitInfoForHeight.Version,
		StoreInfos: storeInfos,
	}

	return commitInfoForHeight, nil
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}
