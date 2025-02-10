package testkv

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"cosmossdk.io/orm/model/ormtable"
)

func NewGoLevelDBBackend(t testing.TB) ormtable.Backend {
	db, err := dbm.NewGoLevelDB("test", t.TempDir(), nil)
	assert.NilError(t, err)
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: db,
	})
}
