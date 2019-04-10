package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}

// Registers types to codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitProposal{}, "params/MsgSubmitParameterChangeProposal", nil)
	cdc.RegisterConcrete(ProposalChange{}, "params/ProposalChange", nil)
}
