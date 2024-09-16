package testkv

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/model/ormtable"
	dbm "cosmossdk.io/store/db"
)

func NewGoLevelDBBackend(tb testing.TB) ormtable.Backend {
	tb.Helper()
	db, err := dbm.NewGoLevelDB("test", tb.TempDir(), nil)
	assert.NilError(tb, err)
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: TestStore{Db: db},
	})
}
