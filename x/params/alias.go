package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/internal/types"
)

var (
	// functions aliases
	NewParameterChangeProposal = types.NewParameterChangeProposal
	NewParamChange             = types.NewParamChange

	// variable aliases
	ModuleCdc = types.ModuleCdc

	RegisterCodec = types.RegisterCodec
	NewCodec      = types.NewCodec
)

type (
	Codec = types.Codec

	ParameterChangeProposal = types.ParameterChangeProposal
	ParamChange             = types.ParamChange
)
