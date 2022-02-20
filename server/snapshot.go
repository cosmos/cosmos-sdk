package server

import (
	"github.com/spf13/cobra"
)

// cmd to create a snapshot
// cmd to restore a snapshot

// RestoreCmd - restore an application.db by using a snapshot
func RestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [height]",
		Short: "restore snapshot stored locally, if no height is picked, restore the latest snapshot",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// find the height specified by the user, if any

			// serverCtx := GetServerContextFromCmd(cmd)
			// cfg := serverCtx.Config

			// nodeKey, err := cfg.LoadNodeKeyID()
			// if err != nil {
			// 	return err
			// }
			// fmt.Println(nodeKey)
			return nil
		},
	}
}
