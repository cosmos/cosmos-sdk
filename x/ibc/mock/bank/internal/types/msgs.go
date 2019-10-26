package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	TypeMsgRecvPacket = "recv_packet"
)

type MsgRecvPacket struct {
	Packet channel.PacketI    `json:"packet" yaml:"packet"`
	Proofs []commitment.Proof `json:"proofs" yaml:"proofs"`
	Height uint64             `json:"height" yaml:"height"`
	Signer sdk.AccAddress     `json:"signer" yaml:"signer"`
}

func NewMsgRecvPacket(packet channel.PacketI, proofs []commitment.Proof, height uint64, signer sdk.AccAddress) MsgRecvPacket {
	return MsgRecvPacket{
		Packet: packet,
		Proofs: proofs,
		Height: height,
		Signer: signer,
	}
}

func (MsgRecvPacket) Route() string {
	return RouterKey
}

func (MsgRecvPacket) Type() string {
	return TypeMsgRecvPacket
}

func (msg MsgRecvPacket) ValidateBasic() sdk.Error {
	if msg.Proofs == nil {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeProofMissing, "proof missing")
	}

	if msg.Signer.Empty() {
		return sdk.NewError(sdk.CodespaceType(DefaultCodespace), CodeInvalidAddress, "invalid signer")
	}

	if err := msg.Packet.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (msg MsgRecvPacket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgRecvPacket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
