package store

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storev2 "cosmossdk.io/store/v2"

	"github.com/cosmos/cosmos-sdk/version"
)

// QueryBlockResultsCmd implements the default command for a BlockResults query.
func (s StoreComponent) PrunesCmd(appCreator serverv2.AppCreator[transaction.Tx]) *cobra.Command {
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
		Example: fmt.Sprintf("%s prune custom --pruning-keep-recent 100 --app-db-backend 'goleveldb'", version.AppName),
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// bind flags to the Context's Viper so we can get pruning options.
			vp := viper.New()
			if err := vp.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := vp.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}

			// use the first argument if present to set the pruning method
			if len(args) > 0 {
				vp.Set(FlagPruning, args[0])
			} else {
				vp.Set(FlagPruning, pruningtypes.PruningOptionDefault)
			}
			pruningOptions, err := getPruningOptionsFromFlags(vp)
			if err != nil {
				return err
			}

			cmd.Printf("get pruning options from command flags, strategy: %v, keep-recent: %v\n",
				pruningOptions.Strategy,
				pruningOptions.KeepRecent,
			)

			logger := log.NewLogger(cmd.OutOrStdout())
			app := appCreator(logger, vp)
			store := app.GetStore()

			rootStore, ok := store.(storev2.RootStore)
			if !ok {
				return fmt.Errorf("currently only support the pruning of rootmulti.Store type")
			}
			latestHeight, err := rootStore.GetLatestVersion()
			if err != nil {
				return err
			}

			// valid heights should be greater than 0.
			if latestHeight <= 0 {
				return fmt.Errorf("the database has no valid heights to prune, the latest height: %v", latestHeight)
			}

			pruningHeight := latestHeight - pruningOptions.KeepRecent
			cmd.Printf("pruning heights up to %v\n", pruningHeight)

			err = rootStore.Prune(pruningHeight)
			if err != nil {
				return err
			}

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

func getPruningOptionsFromFlags(v *viper.Viper) (pruningtypes.PruningOptions, error) {
	strategy := strings.ToLower(cast.ToString(v.Get(FlagPruning)))

	switch strategy {
	case pruningtypes.PruningOptionDefault, pruningtypes.PruningOptionNothing, pruningtypes.PruningOptionEverything:
		return pruningtypes.NewPruningOptionsFromString(strategy), nil

	case pruningtypes.PruningOptionCustom:
		opts := pruningtypes.NewCustomPruningOptions(
			cast.ToUint64(v.Get(FlagPruningKeepRecent)),
			cast.ToUint64(v.Get(FlagPruningInterval)),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return pruningtypes.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
