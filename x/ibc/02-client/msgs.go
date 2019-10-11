package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgCreateClient struct {
	ClientID       string
	ConsensusState ConsensusState
	Signer         sdk.AccAddress
}

func NewMsgCreateClient(clientID string, consState ConsensusState, signer sdk.AccAddress) MsgCreateClient {
	return MsgCreateClient{
		ClientID: clientID,
		ConsensusState: consState,
		Signer: signer,
	}
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
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgUpdateClient struct {
	ClientID string
	Header   Header
	Signer   sdk.AccAddress
}

func NewMsgUpdateClient(cid string, h Header, s sdk.AccAddress) MsgUpdateClient {
	return MsgUpdateClient{
		ClientID: cid,
		Header: h,
		Signer: s,
	}
}

var _ sdk.Msg = MsgUpdateClient{}

func (msg MsgUpdateClient) Route() string {
	return "ibc"
}

func (msg MsgUpdateClient) Type() string {
	return "update-client"
}

func (msg MsgUpdateClient) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("empty address")
	}
	return nil
}

func (msg MsgUpdateClient) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgUpdateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
