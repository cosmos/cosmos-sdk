package server

import (
	"path/filepath"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
)

func TestOpenDB(t *testing.T) {
	t.Parallel()
	_, err := openDB(t.TempDir(), dbm.GoLevelDBBackend)
	require.NoError(t, err)
}

func TestOpenTraceWriter(t *testing.T) {
	t.Parallel()

	fname := filepath.Join(t.TempDir(), "logfile")
	w, err := openTraceWriter(fname)
	require.NoError(t, err)
	require.NotNil(t, w)

	// test no-op
	w, err = openTraceWriter("")
	require.NoError(t, err)
	require.Nil(t, w)
}
