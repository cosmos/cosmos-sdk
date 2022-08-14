package server

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_openDB(t *testing.T) {
	t.Parallel()
	_, err := openDB(t.TempDir())
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
