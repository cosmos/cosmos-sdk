package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Port struct {
	channel Manager
	id      string
	valid   *bool // once invalid forever invalid
}

// bindPort, expected to be called only at init time
// TODO: is it safe to support runtime bindPort?
func (man Manager) Port(id string) Port {
	if _, ok := man.ports[id]; ok {
		panic("port already occupied")
	}
	man.ports[id] = struct{}{}
	valid := true
	return Port{man, id, &valid}
}

// releasePort
func (port Port) Release() {
	delete(port.channel.ports, port.id)
	*port.valid = false
}

func (port Port) Send(ctx sdk.Context, chanid string, packet Packet) error {
	return port.channel.Send(ctx, port.id, chanid, packet)
}

func (port Port) Receive(ctx sdk.Context, proof []commitment.Proof, chanid string, packet Packet) error {
	return port.channel.Receive(ctx, proof, port.id, chanid, packet)
}
