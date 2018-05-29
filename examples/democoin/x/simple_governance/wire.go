package simple_governance

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
	cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)
}
