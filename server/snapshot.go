package server

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	svrtypes "github.com/cosmos/cosmos-sdk/server/types"
)

// RestoreCmd - restore an application.db by using a snapshot
func SnapshotCmd(ac svrtypes.AppCreator) *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot",
		Short: "snapshot the highest block height",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// find the height specified by the user, if any
			serverCtx := GetServerContextFromCmd(cmd)

			db, err := UnsafeOpenDB(serverCtx.Config.RootDir)
			if err != nil {
				return err
			}

			app := ac(serverCtx.Logger, db, nil, serverCtx.Viper)

			debugApp, ok := app.(svrtypes.DebugApp)
			if !ok {
				return errors.New("application does not implement the DebugApp interface")
			}

			sm, err := debugApp.GetBaseApp().SnapshotManager()
			if err != nil {
				return nil
			}

			sm.Create(uint64(debugApp.GetBaseApp().LastBlockHeight()))

			return nil
		},
	}
}

// RestoreCmd - restore an application.db by using a snapshot
func RestoreCmd(ac svrtypes.AppCreator) *cobra.Command {
	return &cobra.Command{
		Use:   "restore ",
		Short: "restore snapshot stored locally, if no height is picked, restore the latest snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {

			// find the height specified by the user, if any
			serverCtx := GetServerContextFromCmd(cmd)

			db, err := UnsafeOpenDB(serverCtx.Config.RootDir)
			if err != nil {
				return err
			}

			app := ac(serverCtx.Logger, db, nil, serverCtx.Viper)

			debugApp, ok := app.(svrtypes.DebugApp)
			if !ok {
				return errors.New("application does not implement the DebugApp interface")
			}

			sm, err := debugApp.GetBaseApp().SnapshotManager()
			if err != nil {
				return nil
			}

			snapshot, err := sm.GetLatest()
			if err != nil {
				return err
			}

			// possible panic?
			sm.Restore(*snapshot)

			return nil
		},
	}
}

// RestoreCmd - restore an application.db by using a snapshot
func ListSnapshotCmd(ac svrtypes.AppCreator) *cobra.Command {
	return &cobra.Command{
		Use:   "snapshots",
		Short: "lists all snapshots stored locally",
		RunE: func(cmd *cobra.Command, args []string) error {

			// find the height specified by the user, if any
			serverCtx := GetServerContextFromCmd(cmd)

			db, err := UnsafeOpenDB(serverCtx.Config.RootDir)
			if err != nil {
				return err
			}

			app := ac(serverCtx.Logger, db, nil, serverCtx.Viper)

			debugApp, ok := app.(svrtypes.DebugApp)
			if !ok {
				return errors.New("application does not implement the DebugApp interface")
			}

			sm, err := debugApp.GetBaseApp().SnapshotManager()
			if err != nil {
				return err
			}

			sh := []uint64{}
			snapshots, err := sm.List()
			for _, s := range snapshots {
				sh = append(sh, s.Height)
			}

			// only print the snapshot heights
			fmt.Println(sh)

			return nil
		},
	}
}
