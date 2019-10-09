package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Port struct {
	channel Manager
	id      string
}

// bindPort, expected to be called only at init time
// TODO: is it safe to support runtime bindPort?
func (man Manager) Port(id string) Port {
	if _, ok := man.ports[id]; ok {
		panic("port already occupied")
	}
	man.ports[id] = struct{}{}
	return Port{man, id}
}

// releasePort
func (port Port) Release() {
	delete(port.channel.ports, port.id)
}

func (man Manager) IsValid(port Port) bool {
	_, ok := man.ports[port.id]
	return ok
}

func (port Port) Send(ctx sdk.Context, chanid string, packet Packet) sdk.Error {
	if !port.channel.IsValid(port) {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(333), "Port is not in valid state")
	}

	if packet.SenderPort() != port.id {
		panic("Packet sent on wrong port")
	}

	return port.channel.Send(ctx, chanid, packet)
}

func (port Port) Receive(ctx sdk.Context, proof []commitment.Proof, height uint64, chanid string, packet Packet) sdk.Error {
	return port.channel.Receive(ctx, proof, height, port.id, chanid, packet)
}
