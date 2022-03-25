package badgerdb

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/dbtest"
)

func load(t *testing.T, dir string) db.DBConnection {
	d, err := NewDB(dir)
	require.NoError(t, err)
	return d
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

func TestVersionManager(t *testing.T) {
	new := func(vs []uint64) db.VersionSet {
		vmap := map[uint64]uint64{}
		var lastTs uint64
		for _, v := range vs {
			vmap[v] = v
			lastTs = v
		}
		return &versionManager{db.NewVersionManager(vs), vmap, lastTs}
	}
	dbtest.DoTestVersionSet(t, new)
}
