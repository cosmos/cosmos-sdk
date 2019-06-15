package group

import "github.com/cosmos/cosmos-sdk/codec"

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateGroup{}, "group/MsgCreateGroup", nil)
	cdc.RegisterConcrete(Group{}, "group/Group", nil)
	cdc.RegisterConcrete(GroupAccount{}, "group/GroupAccount", nil)
	cdc.RegisterConcrete(MsgCreateProposal{}, "group/MsgCreateProposal", nil)
	cdc.RegisterConcrete(MsgVote{}, "group/MsgVote", nil)
	cdc.RegisterConcrete(MsgTryExecuteProposal{}, "group/MsgTryExecuteProposal", nil)
	cdc.RegisterConcrete(MsgWithdrawProposal{}, "group/MsgWithdrawProposal", nil)
	cdc.RegisterConcrete(Proposal{}, "group/Proposal", nil)
}
