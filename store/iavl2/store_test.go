package iavl2

import (
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/types"
)

func TestStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "iavl2")
	require.NoError(t, err, "failed to create temp directory")
	defer os.RemoveAll(dir) // Clean up the temporary directory after the test
	st, err := LoadStore(Config{
		Path: dir,
	}, Options{
		Metrics: metrics.NoOpMetrics{},
		Logger:  log.NewTestLogger(t),
		Key:     types.NewKVStoreKey("test"),
	})
	require.NoError(t, err, "failed to load store")
	var k1, v1 = []byte("key1"), []byte("value1")
	st.Set(k1, v1)
	commit := st.Commit()
	t.Logf("Commit ID: %x, Version: %d", commit.Hash, commit.Version)
	require.Equal(t, st.Get(k1), v1, "expected value to be set correctly")
	st.Delete(k1)
	commit = st.Commit()
	t.Logf("Commit ID: %x, Version: %d", commit.Hash, commit.Version)
	require.Nil(t, st.Get(k1), "expected value to be deleted correctly")
}
