package store

import (
	"testing"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestPrunesCmd(t *testing.T) {
	srv := &Server[transaction.Tx]{}
	cmd := srv.PrunesCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "prune [pruning-method]", cmd.Use)
}

func TestCreateRootStore(t *testing.T) {
	v := viper.New()
	v.Set(FlagAppDBBackend, "goleveldb")
	v.Set("home", t.TempDir())

	logger := log.NewTestLogger(t)
	store, _, err := createRootStore(v, logger, "test")
	require.NoError(t, err)
	require.NotNil(t, store)
}
