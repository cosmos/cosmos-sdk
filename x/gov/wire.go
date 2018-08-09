package gov

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {

	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/MsgVote", nil)
	cdc.RegisterConcrete(DepositProcedure{},"cosmos-sdk/DepositProcedure",nil)
	cdc.RegisterConcrete(TallyingProcedure{},"cosmos-sdk/TallyingProcedure",nil)
	cdc.RegisterConcrete(VotingProcedure{},"cosmos-sdk/VotingProcedure",nil)


	cdc.RegisterInterface((*Proposal)(nil), nil)
	cdc.RegisterConcrete(&TextProposal{}, "gov/TextProposal", nil)
	cdc.RegisterConcrete(&ParameterProposal{}, "gov/ParameterProposal", nil)
	cdc.RegisterConcrete(&SoftwareUpgradeProposal{}, "gov/SoftwareUpgradeProposal", nil)
}

var msgCdc = wire.NewCodec()

func init() {
	RegisterWire(msgCdc)
}
