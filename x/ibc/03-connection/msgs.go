package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const Route = "ibc"

type MsgOpenInit struct {
	ConnectionID       string         `json:"connection_id"`
	Connection         Connection     `json:"connection"`
	CounterpartyClient string         `json:"counterparty_client"`
	NextTimeout        uint64         `json:"next_timeout"`
	Signer             sdk.AccAddress `json:"signer"`
}

var _ sdk.Msg = MsgOpenInit{}

func (msg MsgOpenInit) Route() string {
	return Route
}

func (msg MsgOpenInit) Type() string {
	return "open-init"
}

func (msg MsgOpenInit) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenInit) GetSignBytes() []byte {
	return sdk.MustSortJSON(MsgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenInit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenTry struct {
	ConnectionID       string             `json:"connection_id"`
	Connection         Connection         `json:"connection"`
	CounterpartyClient string             `json:"counterparty_client"`
	Timeout            uint64             `json:"timeout"`
	NextTimeout        uint64             `json:"next_timeout"`
	Proofs             []commitment.Proof `json:"proofs"`
	Signer             sdk.AccAddress     `json:"signer"`
}

var _ sdk.Msg = MsgOpenTry{}

func (msg MsgOpenTry) Route() string {
	return Route
}

func (msg MsgOpenTry) Type() string {
	return "open-try"
}

func (msg MsgOpenTry) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenTry) GetSignBytes() []byte {
	return sdk.MustSortJSON(MsgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenTry) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenAck struct {
	ConnectionID string             `json:"connection_id"`
	Timeout      uint64             `json:"timeout"`
	NextTimeout  uint64             `json:"next_timeout"`
	Proofs       []commitment.Proof `json:"proofs"`
	Signer       sdk.AccAddress     `json:"signer"`
}

var _ sdk.Msg = MsgOpenAck{}

func (msg MsgOpenAck) Route() string {
	return Route
}

func (msg MsgOpenAck) Type() string {
	return "open-ack"
}

func (msg MsgOpenAck) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(MsgCdc.MustMarshalJSON(msg))
}

func (msg MsgOpenAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgOpenConfirm struct {
	ConnectionID string             `json:"connection_id"`
	Timeout      uint64             `json:"timeout"`
	Proofs       []commitment.Proof `json:"proofs"`
	Signer       sdk.AccAddress     `json:"signer"`
}

var _ sdk.Msg = MsgOpenConfirm{}

func (msg MsgOpenConfirm) Route() string {
	return Route
}

func (msg MsgOpenConfirm) Type() string {
	return "open-confirm"
}

func (msg MsgOpenConfirm) ValidateBasic() sdk.Error {
	return nil // TODO
}

func (msg MsgOpenConfirm) GetSignBytes() []byte {
	return nil // TODO
}

func (msg MsgOpenConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
