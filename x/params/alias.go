package params

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/params/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/internal/types"
)

const (
	StoreKey           = keeper.StoreKey
	TStoreKey          = keeper.TStoreKey
	ModuleName         = types.ModuleName
	RouterKey          = types.RouterKey
	ProposalTypeChange = types.ProposalTypeChange
)

var (
	NewKeeper = keeper.NewKeeper

	// functions aliases
	NewParamSetPair            = keeper.NewParamSetPair
	NewSubspace                = keeper.NewSubspace
	NewKeyTable                = keeper.NewKeyTable
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
	ParamSetPair            = keeper.ParamSetPair
	ParamSetPairs           = keeper.ParamSetPairs
	ParamSet                = keeper.ParamSet
	Subspace                = keeper.Subspace
	ReadOnlySubspace        = keeper.ReadOnlySubspace
	KeyTable                = keeper.KeyTable
	ParameterChangeProposal = types.ParameterChangeProposal
	ParamChange             = types.ParamChange
)
