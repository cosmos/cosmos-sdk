package gov

import (
	"github.com/YunSuk-Yeo/cosmos-sdk/codec"
)

var msgCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/MsgVote", nil)

	cdc.RegisterInterface((*ProposalContent)(nil), nil)
	cdc.RegisterConcrete(TextProposal{}, "gov/TextProposal", nil)
	cdc.RegisterConcrete(SoftwareUpgradeProposal{}, "gov/SoftwareUpgradeProposal", nil)
}

func init() {
	RegisterCodec(msgCdc)
}

// SetMsgCodex allows sdk users use custom codex at GetSignBytes
func SetMsgCodec(cdc *codec.Codec) {
	msgCdc = cdc
}
