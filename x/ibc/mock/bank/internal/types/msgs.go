package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgTransfer struct {
	SrcPort    string
	SrcChannel string
	DstPort    string
	DstChannel string
	Amount     sdk.Coin
	Sender     sdk.AccAddress
	Receiver   sdk.AccAddress
	Source     bool
}

func NewMsgTransfer(srcPort, srcChannel, dstPort, dstChannel string, amount sdk.Coin, sender, receiver sdk.AccAddress, source bool) MsgTransfer {
	return MsgTransfer{
		SrcPort:    srcPort,
		SrcChannel: srcChannel,
		DstPort:    dstPort,
		DstChannel: dstChannel,
		Amount:     amount,
		Sender:     sender,
		Receiver:   receiver,
		Source:     source,
	}
}

func (MsgTransfer) Route() string {
	return RouterKey
}

func (MsgTransfer) Type() string {
	return "transfer"
}

func (msg MsgTransfer) ValidateBasic() sdk.Error {
	if !msg.Amount.IsValid() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAmount, "invalid amount")
	}

	if msg.Sender.Empty() || msg.Receiver.Empty() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid address")
	}

	return nil
}

func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(MouduleCdc.MustMarshalJSON(msg))
}

func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
