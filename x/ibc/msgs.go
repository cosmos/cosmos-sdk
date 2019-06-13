package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type MsgCreateClient struct {
	ClientID       string
	ConsensusState ConsensusState
	Signer         sdk.AccAddress
}

type MsgOpenConnection struct {
}

type MsgOpenChannel struct {
}

type MsgReceive struct {
	ConnectionID string
	ChannelID    string
	Packet       channel.Packet
	Signer       sdk.AccAddress
}

var _ sdk.Msg = MsgReceive{}

func (msg MsgReceive) Route() string {
	return "ibc"
}

func (msg MsgReceive) Type() string {
	return "receive"
}

func (msg MsgReceive) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("empty address")
	}
	return nil
}

func (msg MsgReceive) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgReceive) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
