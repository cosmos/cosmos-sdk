package testkv

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

func NewGoLevelDBBackend(t testing.TB) ormtable.Backend {
	db, err := dbm.NewDB("test", "goleveldb", t.TempDir())
	assert.NilError(t, err)
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: db,
	})
}
