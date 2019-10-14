package channel

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

var _ types.StoreKey = (*StoreKey)(nil)

type StoreKey struct {
	name string
}

func NewStoreKey(name string) *StoreKey {
	return &StoreKey{
		name: name,
	}
}

func (key *StoreKey) Name() string {
	return key.name
}

func (key *StoreKey) String() string {
	return fmt.Sprintf("port.StoreKey{%p, %s}", key, key.name)
}

type Port struct {
	channel Manager
	key     *StoreKey
}

// bindPort, expected to be called only at init time
// TODO: is it safe to support runtime bindPort?
func (man Manager) Port(key *StoreKey) Port {
	if key == nil {
		panic("Port() key cannot be nil")
	}
	if _, ok := man.ports[key]; ok {
		panic(fmt.Sprintf("Port duplicate store key %v", key))
	}
	if _, ok := man.portKeyByName[key.Name()]; ok {
		panic(fmt.Sprintf("Port duplicate store key name %v", key))
	}
	man.portKeyByName[key.Name()] = key
	man.ports[key] = struct{}{}
	return Port{man, key}
}

// releasePort
func (port Port) Release() {
	delete(port.channel.portKeyByName, port.key.Name())
	delete(port.channel.ports, port.key)
}

func (man Manager) IsValid(port Port) bool {
	_, ok := man.ports[port.key]
	return ok
}

func (port Port) Send(ctx sdk.Context, chanid string, packet Packet) sdk.Error {
	if !port.channel.IsValid(port) {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(333), "Port is not in valid state")
	}

	if packet.SenderPort() != port.key.Name() {
		panic("Packet sent on wrong port")
	}

	return port.channel.Send(ctx, chanid, packet)
}

func (port Port) Receive(ctx sdk.Context, proof []commitment.Proof, height uint64, chanid string, packet Packet) sdk.Error {
	return port.channel.Receive(ctx, proof, height, port.key.Name(), chanid, packet)
}
