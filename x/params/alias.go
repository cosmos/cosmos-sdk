package params

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/cosmos-sdk/x/params/internal/keeper"
)

const (
	StoreKey           = subspace.StoreKey
	TStoreKey          = subspace.TStoreKey
	ModuleName         = types.ModuleName
	RouterKey          = types.RouterKey
	ProposalTypeChange = types.ProposalTypeChange
)

var (
	NewKeeper = keeper.NewKeeper

	// functions aliases
	NewParamSetPair            = subspace.NewParamSetPair
	NewSubspace                = subspace.NewSubspace
	NewKeyTable                = subspace.NewKeyTable
	RegisterCodec              = types.RegisterCodec
	ErrUnknownSubspace         = types.ErrUnknownSubspace
	ErrSettingParameter        = types.ErrSettingParameter
	ErrEmptyChanges            = types.ErrEmptyChanges
	ErrEmptySubspace           = types.ErrEmptySubspace
	ErrEmptyKey                = types.ErrEmptyKey
	ErrEmptyValue              = types.ErrEmptyValue
	NewParameterChangeProposal = types.NewParameterChangeProposal
	NewParamChange             = types.NewParamChange
	ValidateChanges            = types.ValidateChanges
	NewCodec                   = types.NewCodec

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper = keeper.Keeper

	Codec                   = types.Codec
	ParamSetPair            = subspace.ParamSetPair
	ParamSetPairs           = subspace.ParamSetPairs
	ParamSet                = subspace.ParamSet
	Subspace                = subspace.Subspace
	ReadOnlySubspace        = subspace.ReadOnlySubspace
	KeyTable                = subspace.KeyTable
	ParameterChangeProposal = types.ParameterChangeProposal
	ParamChange             = types.ParamChange
)
