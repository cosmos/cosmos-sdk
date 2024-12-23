package store

import (
	"testing"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestSnapshotCommands(t *testing.T) {
	srv := &Server[transaction.Tx]{}

	tests := []struct {
		name string
		fn   func() *cobra.Command
	}{
		{"export", srv.ExportSnapshotCmd},
		{"restore", srv.RestoreSnapshotCmd},
		{"list", srv.ListSnapshotsCmd},
		{"delete", srv.DeleteSnapshotCmd},
		{"dump", srv.DumpArchiveCmd},
		{"load", srv.LoadArchiveCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.fn()
			require.NotNil(t, cmd)
		})
	}
}

func TestCreateSnapshotsManager(t *testing.T) {
	tmpDir := t.TempDir()
	v := viper.New()
	v.Set("home", tmpDir)

	cmd := &cobra.Command{}
	cmd.Flags().Uint64(FlagKeepRecent, 5, "")
	cmd.Flags().Uint64(FlagInterval, 100, "")

	logger := log.NewTestLogger(t)

	rootStore, _, err := createRootStore(v, logger, "test")
	require.NoError(t, err)

	_, err = createSnapshotsManager(cmd, v, logger, nil)
	require.ErrorContains(t, err, "store is nil") // Should error without valid store backend

	_, err = createSnapshotsManager(cmd, v, logger, rootStore)
	require.NoError(t, err)
}
