package gov

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// TODO: Delete?
func RegisterWire(cdc *wire.Codec) {

	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/MsgVote", nil)
}
