package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/db"
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

			rootStore, keepRecent, err := createRootStore(cmd, vp, logger)
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

			diff := latestHeight - keepRecent
			cmd.Printf("pruning heights up to %v\n", diff)

			err = rootStore.Prune(latestHeight)
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

func createRootStore(cmd *cobra.Command, v *viper.Viper, logger log.Logger) (storev2.RootStore, uint64, error) {
	tempViper := v
	rootDir := v.GetString(serverv2.FlagHome)
	// handle FlagAppDBBackend
	var dbType db.DBType
	if cmd.Flags().Changed(FlagAppDBBackend) {
		dbStr, err := cmd.Flags().GetString(FlagAppDBBackend)
		if err != nil {
			return nil, 0, err
		}
		dbType = db.DBType(dbStr)
	} else {
		dbType = db.DBType(v.GetString(FlagAppDBBackend))
	}
	scRawDb, err := db.NewDB(dbType, "application", filepath.Join(rootDir, "data"), nil)
	if err != nil {
		panic(err)
	}

	// handle KeepRecent & Interval flags
	if cmd.Flags().Changed(FlagKeepRecent) {
		keepRecent, err := cmd.Flags().GetUint64(FlagKeepRecent)
		if err != nil {
			return nil, 0, err
		}

		// viper has an issue that we could not override subitem then Unmarshal key
		// so we can not do viper.Set() as comment below
		// https://github.com/spf13/viper/issues/1106
		// Do it by a hacky: overwrite app.toml file then read config again.

		// v.Set("store.options.sc-pruning-option.keep-recent", keepRecent) // entry that read from app.toml
		// v.Set("store.options.ss-pruning-option.keep-recent", keepRecent)

		err = overrideKeepRecent(filepath.Join(rootDir, "config"), keepRecent)
		if err != nil {
			return nil, 0, err
		}

		tempViper, err = serverv2.ReadConfig(filepath.Join(rootDir, "config"))
		if err != nil {
			return nil, 0, err
		}
	}

	storeOpts := root.DefaultStoreOptions()
	if v != nil && v.Sub("store.options") != nil {
		if err := v.Sub("store.options").Unmarshal(&storeOpts); err != nil {
			return nil, 0, fmt.Errorf("failed to store options: %w", err)
		}
	}

	store, err := root.CreateRootStore(&root.FactoryOptions{
		Logger:  logger,
		RootDir: rootDir,
		Options: storeOpts,
		SCRawDB: scRawDb,
	})

	return store, tempViper.GetUint64("store.options.sc-pruning-option.keep-recent"), err
}

func overrideKeepRecent(configPath string, keepRecent uint64) error {
	bz, err := os.ReadFile(filepath.Join(configPath, "app.toml"))
	if err != nil {
		return err
	}
	lines := strings.Split(string(bz), "\n")

	for i, line := range lines {
		if strings.Contains(line, "keep-recent") {
			lines[i] = fmt.Sprintf("keep-recent = %d", keepRecent)
		}
	}
	output := strings.Join(lines, "\n")

	return os.WriteFile(filepath.Join(configPath, "app.toml"), []byte(output), 0o600)
}
