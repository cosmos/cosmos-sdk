package testkv

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/model/ormtable"
)

func NewGoLevelDBBackend(tb testing.TB) ormtable.Backend {
	tb.Helper()
	db, err := dbm.NewGoLevelDB("test", tb.TempDir(), nil)
	assert.NilError(tb, err)
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: TestStore{Db: db},
	})
}
