package store

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/root"
)

// PrunesCmd implements the default command for pruning app history states.
func (s *Server[T]) PrunesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune [pruning-method]",
		Short: "Prune app history states by keeping the recent heights and deleting old heights",
		Long: `Prune app history states by keeping the recent heights and deleting old heights.
The pruning option is provided via the 'pruning' argument or alternatively with '--pruning-keep-recent'

- default: the last 362880 states are kept
- nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
- everything: 2 latest states will be kept
- custom: allow pruning options to be manually specified through 'pruning-keep-recent'

Note: When the --app-db-backend flag is not specified, the default backend type is 'goleveldb'.
Supported app-db-backend types include 'goleveldb', 'pebbledb'.`,
		Example: "<appd> prune custom --pruning-keep-recent 100 --app-db-backend 'goleveldb'",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// bind flags to the Context's Viper so we can get pruning options.
			vp := serverv2.GetViperFromCmd(cmd)
			if err := vp.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := vp.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}

			latestHeight, err := s.store.GetLatestVersion()
			if err != nil {
				return err
			}

			// valid heights should be greater than 0.
			if latestHeight <= 0 {
				return fmt.Errorf("the database has no valid heights to prune, the latest height: %v", latestHeight)
			}

			err = s.store.Prune(latestHeight)
			if err != nil {
				return err
			}

			cmd.Println("successfully pruned the application root multi stores")
			return nil
		},
	}

	cmd.Flags().String(FlagAppDBBackend, "", "The type of database for application and snapshots databases")
	cmd.Flags().Uint64(FlagKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")

	return cmd
}

// ModuleHashByHeightQuery retrieves the module hashes at a given height.
func (s *Server[T]) ModuleHashByHeightQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "module-hash-by-height <height>",
		Short:   "Get module hashes at a given height",
		Long:    "Get module hashes at a given height. This command is useful for debugging and verifying the state of the application at a given height. Daemon should not be running when calling this command.",
		Example: "<appd module-hash-by-height 16841115",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vp := serverv2.GetViperFromCmd(cmd)
			if err := vp.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := vp.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}

			heightToRetrieveString := args[0]
			height, err := strconv.ParseInt(heightToRetrieveString, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}

			commitInfoForHeight, err := getModuleHashesAtHeight(vp, serverv2.GetLoggerFromCmd(cmd), uint64(height))
			if err != nil {
				return err
			}

			bytes, err := json.Marshal(commitInfoForHeight)
			if err != nil {
				return fmt.Errorf("failed to marshal commit info: %w", err)
			}

			cmd.Println(string(bytes))
			return nil
		},
	}

	return cmd
}

func getModuleHashesAtHeight(vp *viper.Viper, logger log.Logger, height uint64) (*proof.CommitInfo, error) {
	rootStore, _, err := createRootStore(vp, logger)
	if err != nil {
		return nil, fmt.Errorf("can not create root store %w", err)
	}

	commitInfoForHeight, err := rootStore.GetStateCommitment().GetCommitInfo(height)
	if err != nil {
		return nil, err
	}

	// Create a new slice of StoreInfos for storing the modified hashes.
	storeInfos := make([]*proof.StoreInfo, len(commitInfoForHeight.StoreInfos))

	for i, storeInfo := range commitInfoForHeight.StoreInfos {
		// Convert the hash to a hexadecimal string.
		hash := strings.ToUpper(hex.EncodeToString(storeInfo.CommitId.Hash))

		// Create a new StoreInfo with the modified hash.
		storeInfos[i] = &proof.StoreInfo{
			Name: storeInfo.Name,
			CommitId: &proof.CommitID{
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
	commitInfoForHeight = &proof.CommitInfo{
		Version:    commitInfoForHeight.Version,
		StoreInfos: storeInfos,
		Timestamp:  commitInfoForHeight.Timestamp,
	}

	return commitInfoForHeight, nil
}

func createRootStore(v *viper.Viper, logger log.Logger) (storev2.RootStore, root.Options, error) {
	storeConfig, err := UnmarshalConfig(v.AllSettings())
	if err != nil {
		return nil, root.Options{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	store, err := root.NewBuilder().Build(logger, storeConfig)
	if err != nil {
		return nil, root.Options{}, fmt.Errorf("failed to create store backend: %w", err)
	}
	return store, storeConfig.Options, nil
}
