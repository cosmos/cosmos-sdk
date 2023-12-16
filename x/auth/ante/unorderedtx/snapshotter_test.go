package unorderedtx_test

import (
	"testing"

	"cosmossdk.io/x/auth/ante/unorderedtx"
	"github.com/stretchr/testify/require"
)

func TestSnapshotter_SnapshotExtension(t *testing.T) {
	dataDir := t.TempDir()
	txm := unorderedtx.NewManager(dataDir)
	txm.Start()

	defer txm.Close()

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
}

func TestSnapshotter_RestoreExtension(t *testing.T) {
}
