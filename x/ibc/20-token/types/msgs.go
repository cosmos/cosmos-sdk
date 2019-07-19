package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgSend struct {
	FromAddress  sdk.AccAddress `json:"from_address"`
	ToAddress    sdk.AccAddress `json:"to_address"`
	ToConnection string         `json:"to_connection"`
	ToChannel    string         `json:"to_channel"`
	Amount       sdk.Coins      `json:"amount"`
}

var _ sdk.Msg = MsgSend{}

func NewMsgSend(fromAddr, toAddr sdk.AccAddress, toConnection, toChannel string, amount sdk.Coins) MsgSend {
	return MsgSend{
		FromAddress:  fromAddr,
		ToAddress:    toAddr,
		ToConnection: toConnection,
		ToChannel:    toChannel,
		Amount:       amount,
	}
}

func (msg MsgSend) Route() string {
	return "token"
}

func (msg MsgSend) Type() string {
	return "send"
}

func (msg MsgSend) ValidateBasic() sdk.Error {
	if msg.FromAddress.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if msg.ToAddress.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("send amount is invalid")
	}
	if !msg.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("send amount must be positive")
	}
	return nil
}

func (msg MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleCdc.MustMarshalJSON(msg))
}

func (msg MsgSend) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.FromAddress}
}
