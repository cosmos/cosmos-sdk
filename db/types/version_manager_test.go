package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/db/dbtest"
	"github.com/cosmos/cosmos-sdk/db/types"
)

func TestVersionManager(t *testing.T) {
	new := func(vs []uint64) types.VersionSet { return types.NewVersionManager(vs) }
	dbtest.DoTestVersionSet(t, new)
}
