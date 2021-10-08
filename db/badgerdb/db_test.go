package badgerdb

import (
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
