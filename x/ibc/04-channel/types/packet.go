package types

import (
	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
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
	data PacketData,
	sequence uint64, sourcePort, sourceChannel,
	destinationPort, destinationChannel string,
) Packet {
	return Packet{
		Data:               data,
		Sequence:           sequence,
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    destinationPort,
		DestinationChannel: destinationChannel,
	}
}

// Type exports Packet.Data.Type
func (p Packet) Type() string {
	return p.Data.Type()
}

// GetTimeoutHeight implements PacketI interface
func (p Packet) GetTimeoutHeight() uint64 { return p.Data.GetTimeoutHeight() }

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
	if p.Data.GetTimeoutHeight() == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet timeout cannot be 0")
	}
	if len(p.Data.GetBytes()) == 0 {
		return sdkerrors.Wrap(ErrInvalidPacket, "packet data bytes cannot be empty")
	}
	return p.Data.ValidateBasic()
}
