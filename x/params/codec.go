package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Registers types to codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitParameterChangeProposal{}, "params/MsgSubmitParameterChangeProposal", nil)

	cdc.RegisterConcrete(ProposalChange{}, "params/ProposalChange", nil)
}
