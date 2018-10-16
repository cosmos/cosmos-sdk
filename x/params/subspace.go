package params

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// re-export types from subspace
type (
	Subspace         = subspace.Subspace
	ReadOnlySubspace = subspace.ReadOnlySubspace
	ParamSet         = subspace.ParamSet
	KeyValuePairs    = subspace.KeyValuePairs
	TypeTable        = subspace.TypeTable
)

// re-export functions from subspace
func NewTypeTable(keytypes ...interface{}) TypeTable {
	return subspace.NewTypeTable(keytypes...)
}
func DefaultTestComponents(t *testing.T, table TypeTable) (sdk.Context, Subspace, func() sdk.CommitID) {
	return subspace.DefaultTestComponents(t, table)
}
