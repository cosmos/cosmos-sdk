package testkv

import (
	"testing"

	"gotest.tools/v3/assert"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/orm/model/ormtable"
)

func NewGoLevelDBBackend(tb testing.TB) ormtable.Backend {
	tb.Helper()
	db, err := coretesting.NewGoLevelDB("test", tb.TempDir(), nil)
	assert.NilError(tb, err)
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: TestStore{Db: db},
	})
}
