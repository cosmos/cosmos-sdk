package simpleGovernance

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterWire registers messages into the wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
	cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)
}
