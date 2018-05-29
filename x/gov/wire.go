package gov

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/SubmitProposalMsg", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/DepositMsg", nil)
	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/VoteMsg", nil)

}
