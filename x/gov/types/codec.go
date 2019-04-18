package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	MsgCdc = codec.New()
)

// RegisterCodec registers all the necessary types and interfaces for
// governance.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Content)(nil), nil)

	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/MsgVote", nil)

	cdc.RegisterConcrete(TextProposal{}, "cosmos-sdk/TextProposal", nil)
	cdc.RegisterConcrete(SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal", nil)
}

func init() {
	RegisterCodec(MsgCdc)
}
