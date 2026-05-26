package server

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
)

func Test_openDB(t *testing.T) {
	t.Parallel()
	_, err := openDB(t.TempDir(), dbm.GoLevelDBBackend)
	require.NoError(t, err)
}
