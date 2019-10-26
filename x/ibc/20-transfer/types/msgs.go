package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgTransfer = "transfer"
)

type MsgTransfer struct {
	SrcPort      string         `json:"src_port" yaml:"src_port"`
	SrcChannel   string         `json:"src_channel" yaml:"src_channel"`
	Denomination string         `json:"denomination" yaml:"denomination"`
	Amount       sdk.Int        `json:"amount" yaml:"amount"`
	Sender       sdk.AccAddress `json:"sender" yaml:"sender"`
	Receiver     string         `json:"receiver" yaml:"receiver"`
	Source       bool           `json:"source" yaml:"source"`
}

func NewMsgTransfer(srcPort, srcChannel string, denom string, amount sdk.Int, sender sdk.AccAddress, receiver string, source bool) MsgTransfer {
	return MsgTransfer{
		SrcPort:      srcPort,
		SrcChannel:   srcChannel,
		Denomination: denom,
		Amount:       amount,
		Sender:       sender,
		Receiver:     receiver,
		Source:       source,
	}
}

func (MsgTransfer) Route() string {
	return RouterKey
}

func (MsgTransfer) Type() string {
	return TypeMsgTransfer
}

func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if !msg.Amount.IsPositive() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAmount, "invalid amount")
	}

	if msg.Sender.Empty() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid address")
	}

	if len(msg.Receiver) == 0 {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidReceiver, "receiver is empty")
	}

	return nil
}

func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
