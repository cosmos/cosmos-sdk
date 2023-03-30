package testkv

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	"gotest.tools/v3/assert"
)

func NewGoLevelDBBackend(tb testing.TB) ormtable.Backend {
	tb.Helper()
	db, err := dbm.NewGoLevelDB("test", tb.TempDir(), nil)
	assert.NilError(tb, err)
	return ormtable.NewBackend(ormtable.BackendOptions{
		CommitmentStore: db,
	})
}
