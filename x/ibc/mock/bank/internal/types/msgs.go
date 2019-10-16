package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type MsgTransfer struct {
	SrcPort     string
	SrcChannel  string
	Amount      sdk.Coin
	Sender      sdk.AccAddress
	Receiver    string
	Source      bool
	Timeout     uint64
	Proof       ics23.Proof
	ProofHeight uint64
}

func NewMsgTransfer(srcPort, srcChannel string, amount sdk.Coin, sender sdk.AccAddress, receiver string, source bool, timeout uint64, proof ics23.Proof, proofHeight uint64) MsgTransfer {
	return MsgTransfer{
		SrcPort:     srcPort,
		SrcChannel:  srcChannel,
		Amount:      amount,
		Sender:      sender,
		Receiver:    receiver,
		Source:      source,
		Timeout:     timeout,
		Proof:       proof,
		ProofHeight: proofHeight,
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

	if msg.Sender.Empty() || len(msg.Receiver) == 0 {
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
