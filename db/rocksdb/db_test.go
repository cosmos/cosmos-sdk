package rocksdb

import (
	"os"
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
	_, err = db.SaveNextVersion()
	require.NoError(t, err)
	txn := db.Writer()
	require.NoError(t, txn.Set([]byte{1}, []byte{1}))
	require.NoError(t, txn.Set([]byte{2}, []byte{2}))
	require.NoError(t, txn.Commit())

	// make checkpoints dir temporarily unreadable to trigger an error
	require.NoError(t, os.Chmod(db.checkpointsDir(), 0000))
	require.Error(t, db.Revert())

	require.NoError(t, os.Chmod(db.checkpointsDir(), 0755))
	db, err = NewDB(dir)
	require.NoError(t, err)
}
