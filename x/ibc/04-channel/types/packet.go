package types

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	transfertypes "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

//CommitPacket
func CommitPacket(data exported.PacketDataI) []byte {
	bz := sdk.Uint64ToBigEndian(data.GetTimeoutHeight())
	return tmhash.Sum(bz)
}

// CommitAcknowledgement returns the hash of commitment bytes
func CommitAcknowledgement(data exported.PacketAcknowledgementI) []byte {
	return tmhash.Sum(data.GetBytes())
}

var _ exported.PacketI = Packet{}

// NewPacket creates a new Packet instance
func NewPacket(
	data exported.PacketDataI,
	sequence uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string,
) Packet {
	var packetData PacketData

	switch cast := data.(type) {
	case transfertypes.FungibleTokenPacketData:
		packetData = PacketData{
			Value: &PacketData_FungibleToken{&cast},
		}
	default: 
		panic(fmt.Sprintf("invalid packet data type %T", cast))
	}
	
	return Packet{
		Data:               packetData,
		Sequence:           sequence,
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    destinationPort,
		DestinationChannel: destinationChannel,
	}
}

// GetSequence implements PacketI interface
func (p Packet) GetSequence() uint64 { return p.Sequence }

// GetSourcePort implements PacketI interface
func (p Packet) GetSourcePort() string { return p.SourcePort }

// GetSourceChannel implements PacketI interface
func (p Packet) GetSourceChannel() string { return p.SourceChannel }

// GetDestPort implements PacketI interface
func (p Packet) GetDestinationPort() string { return p.DestinationPort }

// GetDestChannel implements PacketI interface
func (p Packet) GetDestinationChannel() string { return p.DestinationChannel }

// GetData returns the concrete packet data type as an interface.
func (p Packet) GetData() exported.PacketDataI {
	return p.Data.GetPacketDataI()
}

// Type exports Packet.Data.Type
func (p Packet) Type() string {
	return p.Data.GetPacketDataI().Type()
}

// GetTimeoutHeight implements PacketI interface
func (p Packet) GetTimeoutHeight() uint64 { return p.Data.GetPacketDataI().GetTimeoutHeight() }

// ValidateBasic implements PacketI interface
func (p Packet) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(p.SourcePort); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid source port ID: %s", p.SourcePort).Error(),
		)
	}
	if err := host.DefaultPortIdentifierValidator(p.DestinationPort); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid destination port ID: %s", p.DestinationPort).Error(),
		)
	}
	if err := host.DefaultChannelIdentifierValidator(p.SourceChannel); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid source channel ID: %s", p.SourceChannel).Error(),
		)
	}
	if err := host.DefaultChannelIdentifierValidator(p.DestinationChannel); err != nil {
		return sdkerrors.Wrapf(
			ErrInvalidPacket,
			sdkerrors.Wrapf(err, "invalid destination channel ID: %s", p.DestinationChannel).Error(),
		)
	}
	if p.Sequence == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet sequence cannot be 0")
	}
	if p.Data.GetPacketDataI().GetTimeoutHeight() == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet timeout cannot be 0")
	}
	if len(p.Data.GetPacketDataI().GetBytes()) == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet data bytes cannot be empty")
	}
	return p.Data.GetPacketDataI().ValidateBasic()
}
