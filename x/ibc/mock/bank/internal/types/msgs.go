package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	TypeMsgRecvTransferPacket = "recv-transfer-packet"
)

type MsgRecvTransferPacket struct {
	Packet ics04.PacketI  `json:"packet" yaml:"packet"`
	Proofs []ics23.Proof  `json:"proofs" yaml:"proofs"`
	Height uint64         `json:"height" yaml:"height"`
	Signer sdk.AccAddress `json:"signer" yaml:"signer"`
}

func NewMsgRecvTransferPacket(packet ics04.PacketI, proofs []ics23.Proof, height uint64, signer sdk.AccAddress) MsgRecvTransferPacket {
	return MsgRecvTransferPacket{
		Packet: packet,
		Proofs: proofs,
		Height: height,
		Signer: signer,
	}
}

func (MsgRecvTransferPacket) Route() string {
	return RouterKey
}

func (MsgRecvTransferPacket) Type() string {
	return TypeMsgRecvTransferPacket
}

func (msg MsgRecvTransferPacket) ValidateBasic() sdk.Error {
	if msg.Proofs == nil {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeProofMissing, "proof missing")
	}

	if msg.Signer.Empty() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid signer")
	}

	return nil
}

func (msg MsgRecvTransferPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(MouduleCdc.MustMarshalJSON(msg))
}

func (msg MsgRecvTransferPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
