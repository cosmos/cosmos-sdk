package db_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/dbtest"
)

func TestVersionManager(t *testing.T) {
	new := func(vs []uint64) db.VersionSet { return db.NewVersionManager(vs) }
	dbtest.DoTestVersionSet(t, new)
}
