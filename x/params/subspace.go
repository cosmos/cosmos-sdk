package params

import (
	"testing"

	sdk "github.com/YunSuk-Yeo/cosmos-sdk/types"

	"github.com/YunSuk-Yeo/cosmos-sdk/x/params/subspace"
)

// re-export types from subspace
type (
	Subspace         = subspace.Subspace
	ReadOnlySubspace = subspace.ReadOnlySubspace
	ParamSet         = subspace.ParamSet
	ParamSetPairs    = subspace.ParamSetPairs
	KeyTable         = subspace.KeyTable
)

// nolint - re-export functions from subspace
func NewKeyTable(keytypes ...interface{}) KeyTable {
	return subspace.NewKeyTable(keytypes...)
}
func DefaultTestComponents(t *testing.T) (sdk.Context, Subspace, func() sdk.CommitID) {
	return subspace.DefaultTestComponents(t)
}
