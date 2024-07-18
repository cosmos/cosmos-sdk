package store

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/root"
)

// QueryBlockResultsCmd implements the default command for a BlockResults query.
func (s *StoreComponent[T]) PrunesCmd() *cobra.Command {
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
Supported app-db-backend types include 'goleveldb', 'rocksdb', 'pebbledb'.`,
		Example: fmt.Sprintf("%s prune custom --pruning-keep-recent 100 --app-db-backend 'goleveldb'", "<appd>"),
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

			logger := log.NewLogger(cmd.OutOrStdout())
			home, err := cmd.Flags().GetString(serverv2.FlagHome) // should be FlagHome
			if err != nil {
				return err
			}

			rootStore, err, interval := createRootStore(cmd, home, vp, logger)
			if err != nil {
				return fmt.Errorf("can not create root store %w", err)
			}

			latestHeight, err := rootStore.GetLatestVersion()
			if err != nil {
				return err
			}

			// valid heights should be greater than 0.
			if latestHeight <= 0 {
				return fmt.Errorf("the database has no valid heights to prune, the latest height: %v", latestHeight)
			}

			upTo := latestHeight - interval
			cmd.Printf("pruning heights up to %v\n", upTo)

			before, err := rootStore.GetStateCommitment().GetCommitInfo(2)
			fmt.Println("before", before, err)

			err = rootStore.Prune(latestHeight)
			if err != nil {
				return err
			}

			after, err := rootStore.GetStateCommitment().GetCommitInfo(2)
			fmt.Println("after", after, err)

			cmd.Println("successfully pruned the application root multi stores")
			return nil
		},
	}

	cmd.Flags().String(FlagAppDBBackend, "", "The type of database for application and snapshots databases")
	cmd.Flags().Uint64(FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningInterval, 10,
		`Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom'), 
		this is not used by this command but kept for compatibility with the complete pruning options`)

	return cmd
}

func createRootStore(cmd *cobra.Command, rootDir string, v *viper.Viper, logger log.Logger) (storev2.RootStore, error, uint64) {
	scRawDb, err := db.NewGoLevelDB("application", filepath.Join(rootDir, "data"), nil)
	if err != nil {
		panic(err)
	}

	// handle KeepRecent & Interval flags
	if cmd.Flags().Changed(FlagPruningKeepRecent) {
		keepRecent, err := cmd.Flags().GetUint64(FlagPruningKeepRecent)
		if err != nil {
			return nil, err, 0
		}

		// Expect ss & sc have same pruning options
		viper.Set("store.options.sc-pruning-option.keep-recent", keepRecent) // entry that read from app.toml
		viper.Set("store.options.ss-pruning-option.keep-recent", keepRecent)
	}

	if cmd.Flags().Changed(FlagPruningInterval) {
		interval, err := cmd.Flags().GetUint64(FlagPruningInterval)
		if err != nil {
			return nil, err, 0
		}

		viper.Set("store.options.sc-pruning-option.interval", interval)
		viper.Set("store.options.ss-pruning-option.interval", interval)
	}

	store, err := root.CreateRootStore(&root.FactoryOptions{
		Logger:  logger,
		RootDir: rootDir,
		Options: v,
		SCRawDB: scRawDb,
	})

	return store, err, viper.GetUint64("store.options.sc-pruning-option.keep-recent")
}
