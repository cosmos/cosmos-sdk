package server

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "cosmossdk.io/store/db"
)

func Test_OpenDB(t *testing.T) {
	t.Parallel()
	_, err := OpenDB(t.TempDir(), dbm.DBTypeGoLevelDB)
	require.NoError(t, err)
}

func Test_openTraceWriter(t *testing.T) {
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
