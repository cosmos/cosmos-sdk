package server

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
)

func Test_openDB(t *testing.T) {
	t.Parallel()
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	_, err := openDB(dir)
	require.NoError(t, err)
}

func Test_openTraceWriter(t *testing.T) {
	t.Parallel()
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	fname := filepath.Join(dir, "logfile")
	w, err := openTraceWriter(fname)
	require.NoError(t, err)
	require.NotNil(t, w)

	// test no-op
	w, err = openTraceWriter("")
	require.NoError(t, err)
	require.Nil(t, w)
}
