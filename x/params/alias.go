// nolint
package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultCodespace = types.DefaultCodespace
)

type (
	Subspace         = subspace.Subspace
	ReadOnlySubspace = subspace.ReadOnlySubspace
	ParamSet         = subspace.ParamSet
	ParamSetPairs    = subspace.ParamSetPairs
	KeyTable         = subspace.KeyTable
)

var (
	NewKeyTable           = subspace.NewKeyTable
	DefaultTestComponents = subspace.DefaultTestComponents
)
