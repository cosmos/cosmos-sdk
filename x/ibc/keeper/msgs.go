package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type MsgCreateClient struct {
	ClientID       string
	ConsensusState client.ConsensusState
	Signer         sdk.AccAddress
}

var _ sdk.Msg = MsgCreateClient{}

func (msg MsgCreateClient) Route() string {
	return "ibc"
}

func (msg MsgCreateClient) Type() string {
	return "create-client"
}

func (msg MsgCreateClient) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("empty address")
	}
	return nil
}

func (msg MsgCreateClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenConnection struct {
	ConnectionID         string
	ClientID             string
	CounterpartyID       string
	CounterpartyClientID string
	Signer               sdk.AccAddress
}

func (msg MsgOpenConnection) Route() string {
	return "ibc"
}

func (msg MsgOpenConnection) Type() string {
	return "create-connection"
}

func (msg MsgOpenConnection) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("empty address")
	}
	return nil
}

func (msg MsgOpenConnection) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgOpenConnection) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgReceive struct {
	ConnectionID string
	ChannelID    string
	Packet       channel.Packet
	Proofs       []commitment.Proof
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
