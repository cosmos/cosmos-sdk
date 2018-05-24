package gov

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
		// TODO include option to always include prefix bytes.
	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/SubmitProposalMsg", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/DepositMsg", nil)
	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/VoteMsg", nil)

}
