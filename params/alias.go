package params

// nolint

import (
	"github.com/cosmos/cosmos-sdk/params/manager"
	"github.com/cosmos/cosmos-sdk/params/types"
)

const (
	RouterKey = types.RouterKey
)

var (
	// functions aliases
	RegisterCodec              = types.RegisterCodec
	ErrUnknownSubspace         = types.ErrUnknownSubspace
	ErrSettingParameter        = types.ErrSettingParameter
	NewParameterChangeProposal = types.NewParameterChangeProposal
	NewParamChange             = types.NewParamChange
	NewCodec                   = types.NewCodec

	// variable aliases
	ModuleCdc = types.ModuleCdc
)

type (
	Keeper = manager.Manager

	Codec                   = types.Codec
	ParameterChangeProposal = types.ParameterChangeProposal
	ParamChange             = types.ParamChange
)
