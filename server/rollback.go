package server

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server/types"
	tmcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
)

const FlagDeletePendingBlock = "delete-pending-block"

// NewRollbackCmd creates a command to rollback tendermint and multistore state by one height.
func NewRollbackCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "rollback cosmos-sdk and tendermint state by one height",
		Long: `
A state rollback is performed to recover from an incorrect application state transition,
when Tendermint has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application also roll back to height n - 1. No blocks are removed, so upon
restarting Tendermint the transactions in block n will be re-executed against the
application.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := GetServerContextFromCmd(cmd)
			cfg := ctx.Config

			deletePendingBlock, err := cmd.Flags().GetBool(FlagDeletePendingBlock)
			if err != nil {
				return err
			}

			if deletePendingBlock {
				if err := deletePendingBlockIfExists(cfg); err != nil {
					return err
				}
			}

			home := cfg.RootDir
			db, err := openDB(home)
			if err != nil {
				return err
			}
			app := appCreator(ctx.Logger, db, nil, ctx.Viper)
			// rollback tendermint state
			height, hash, err := tmcmd.RollbackState(ctx.Config)
			if err != nil {
				return fmt.Errorf("failed to rollback tendermint state: %w", err)
			}
			// rollback the multistore
			if err := app.CommitMultiStore().RollbackToVersion(height); err != nil {
				return fmt.Errorf("failed to rollback to version: %w", err)
			}

			fmt.Printf("Rolled back state to height %d and hash %X", height, hash)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().Bool(FlagDeletePendingBlock, false, "Delete the pending block in tendermint block store if exists")
	return cmd
}

func loadStateAndBlockStore(config *cfg.Config) (*store.BlockStore, state.Store, error) {
	dbType := dbm.BackendType(config.DBBackend)

	// Get BlockStore
	blockStoreDB, err := dbm.NewDB("blockstore", dbType, config.DBDir())
	if err != nil {
		return nil, nil, err
	}
	blockStore := store.NewBlockStore(blockStoreDB)

	// Get StateStore
	stateDB, err := dbm.NewDB("state", dbType, config.DBDir())
	if err != nil {
		return nil, nil, err
	}
	stateStore := state.NewStore(stateDB)

	return blockStore, stateStore, nil
}

func deletePendingBlockIfExists(config *cfg.Config) error {
	blockStore, stateStore, err := loadStateAndBlockStore(config)
	if err != nil {
		return err
	}
	defer func() {
		_ = blockStore.Close()
		_ = stateStore.Close()
	}()
	tmState, err := stateStore.Load()
	if err != nil {
		return err
	}
	if tmState.IsEmpty() {
		return errors.New("no state found")
	}

	height := blockStore.Height()
	if height == tmState.LastBlockHeight+1 {
		// delete this pending block
		if err := blockStore.Rollback(); err != nil {
			return err
		}
		fmt.Println("rollback pending block", height)
	}
	return nil
}
