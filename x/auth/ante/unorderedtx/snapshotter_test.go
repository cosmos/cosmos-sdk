package unorderedtx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/auth/ante/unorderedtx"
)

func TestSnapshotter(t *testing.T) {
	dataDir := t.TempDir()
	txm := unorderedtx.NewManager(dataDir)

	// add a handful of unordered txs
	for i := 0; i < 100; i++ {
		txm.Add([32]byte{byte(i)}, 100)
	}

	var unorderedTxBz []byte
	s := unorderedtx.NewSnapshotter(txm)
	w := func(bz []byte) error {
		unorderedTxBz = bz
		return nil
	}

	err := s.SnapshotExtension(50, w)
	require.NoError(t, err)
	require.NotEmpty(t, unorderedTxBz)

	pr := func() ([]byte, error) {
		return unorderedTxBz, nil
	}

	// restore with an invalid format which should result in an error
	err = s.RestoreExtension(50, 2, pr)
	require.Error(t, err)

	// restore with height > ttl which should result in no unordered txs synced
	txm2 := unorderedtx.NewManager(dataDir)
	s2 := unorderedtx.NewSnapshotter(txm2)
	err = s2.RestoreExtension(200, unorderedtx.SnapshotFormat, pr)
	require.NoError(t, err)
	require.Empty(t, txm2.Size())

	// restore with with height < ttl which should result in all unordered txs synced
	txm3 := unorderedtx.NewManager(dataDir)
	s3 := unorderedtx.NewSnapshotter(txm3)
	err = s3.RestoreExtension(50, unorderedtx.SnapshotFormat, pr)
	require.NoError(t, err)
	require.Equal(t, 100, txm3.Size())

	for i := 0; i < 100; i++ {
		require.True(t, txm3.Contains([32]byte{byte(i)}))
	}
}
