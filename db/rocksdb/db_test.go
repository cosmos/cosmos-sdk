//go:build rocksdb_build

package rocksdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/dbtest"
)

func load(t *testing.T, dir string) dbm.DBConnection {
	db, err := NewDB(dir)
	require.NoError(t, err)
	return db
}

func TestGetSetHasDelete(t *testing.T) {
	dbtest.DoTestGetSetHasDelete(t, load)
}

func TestIterators(t *testing.T) {
	dbtest.DoTestIterators(t, load)
}

func TestTransactions(t *testing.T) {
	dbtest.DoTestTransactions(t, load, true)
}

func TestVersioning(t *testing.T) {
	dbtest.DoTestVersioning(t, load)
}

func TestRevert(t *testing.T) {
	dbtest.DoTestRevert(t, load, false)
	dbtest.DoTestRevert(t, load, true)
}

func TestReloadDB(t *testing.T) {
	dbtest.DoTestReloadDB(t, load)
}

// Test that the DB can be reloaded after a failed Revert
func TestRevertRecovery(t *testing.T) {
	dir := t.TempDir()
	db, err := NewDB(dir)
	require.NoError(t, err)
	txn := db.Writer()
	require.NoError(t, txn.Set([]byte{1}, []byte{1}))
	require.NoError(t, txn.Commit())
	_, err = db.SaveNextVersion()
	require.NoError(t, err)
	txn = db.Writer()
	require.NoError(t, txn.Set([]byte{2}, []byte{2}))
	require.NoError(t, txn.Commit())

	// move checkpoints dir temporarily to trigger an error
	hideDir := filepath.Join(dir, "hide_checkpoints")
	require.NoError(t, os.Rename(db.checkpointsDir(), hideDir))
	require.Error(t, db.Revert())
	require.NoError(t, os.Rename(hideDir, db.checkpointsDir()))

	db, err = NewDB(dir)
	require.NoError(t, err)
	view := db.Reader()
	val, err := view.Get([]byte{1})
	require.NoError(t, err)
	require.Equal(t, []byte{1}, val)
	val, err = view.Get([]byte{2})
	require.NoError(t, err)
	require.Nil(t, val)
	view.Discard()
}
