package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var cdc = codec.New()

type MsgSequence struct {
	Sequence  uint64
	ChannelID string
	Signer    sdk.AccAddress
}

func NewMsgSequence(signer sdk.AccAddress, chanid string, sequence uint64) MsgSequence {
	return MsgSequence{
		Sequence:  sequence,
		ChannelID: chanid,
		Signer:    signer,
	}
}

func (MsgSequence) Route() string {
	return "ibcmock"
}

func (MsgSequence) Type() string {
	return "sequence"
}

func (msg MsgSequence) ValidateBasic() sdk.Error {
	return nil
}

func (msg MsgSequence) GetSignBytes() []byte {
	return sdk.MustSortJSON(cdc.MustMarshalJSON(msg))
}

func (msg MsgSequence) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
