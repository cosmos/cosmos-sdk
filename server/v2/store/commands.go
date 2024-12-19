package store

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
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
			logger := serverv2.GetLoggerFromCmd(cmd)

			rootStore, opts, err := createRootStore(vp, logger)
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

			diff := latestHeight - opts.SCPruningOption.KeepRecent
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
