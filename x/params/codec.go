package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitParameterChangeProposal{}, "params/MsgSubmitParameterChangeProposal", nil)

	cdc.RegisterConcrete(ProposalChange{}, "params/ProposalChange", nil)
}
