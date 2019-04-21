// nolint
package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultCodespace = types.DefaultCodespace
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
	TStoreKey  = types.TStoreKey

	ProposalTypeChange = types.ProposalTypeChange
)

type (
	Subspace         = subspace.Subspace
	ReadOnlySubspace = subspace.ReadOnlySubspace
	ParamSet         = subspace.ParamSet
	ParamSetPairs    = subspace.ParamSetPairs
	KeyTable         = subspace.KeyTable

	ParameterChangeProposal = types.ParameterChangeProposal
	ParamChange             = types.ParamChange
)

var (
	NewKeyTable           = subspace.NewKeyTable
	DefaultTestComponents = subspace.DefaultTestComponents

	NewParamChange             = types.NewParamChange
	NewParameterChangeProposal = types.NewParameterChangeProposal

	ErrUnknownSubspace  = types.ErrUnknownSubspace
	ErrSettingParameter = types.ErrSettingParameter
	ErrEmptyChanges     = types.ErrEmptyChanges
	ErrEmptySubspace    = types.ErrEmptySubspace
	ErrEmptyKey         = types.ErrEmptyKey
	ErrEmptyValue       = types.ErrEmptyValue
)
