package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (k Keeper) Send(ctx sdk.Context, channelID string, packet exported.PacketI, port types.Port) error {
	if !port.channel.IsValid(port) {
		return errors.New("Port is not in valid state")
	}

	if packet.SenderPort() != port.ID() {
		panic("Packet sent on wrong port")
	}

	return port.channel.Send(ctx, channelID, packet)
}

func (k Keeper) Receive(ctx sdk.Context, proof []ics23.Proof, height uint64, channelID string, packet exported.PacketI, port types.Port) error {
	return port.channel.Receive(ctx, proof, height, port.ID(), channelID, packet)
}
