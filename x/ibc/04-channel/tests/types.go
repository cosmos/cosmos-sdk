package channel

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/tests"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const PortName = "port-test"

type Node struct {
	*connection.Node
	Counterparty *Node

	Channel channel.Channel

	Cdc *codec.Codec
}

func NewNode(self, counter tendermint.MockValidators, cdc *codec.Codec) *Node {
	res := &Node{
		Node: connection.NewNode(self, counter, cdc),
		Cdc:  cdc,
	}

	res.Counterparty = &Node{
		Node:         res.Node.Counterparty,
		Counterparty: res,
		Cdc:          cdc,
	}

	res.Channel = channel.Channel{
		Counterparty:     res.Counterparty.Name,
		CounterpartyPort: PortName,
		ConnectionHops:   []string{res.Name},
	}

	res.Counterparty.Channel = channel.Channel{
		Counterparty:     res.Name,
		CounterpartyPort: PortName,
		ConnectionHops:   []string{res.Counterparty.Name},
	}

	return res
}

func (node *Node) Handshaker(t *testing.T, proofs []commitment.Proof) (sdk.Context, channel.Handshaker) {
	ctx := node.Context()
	store, err := commitment.NewStore(node.Counterparty.Root(), node.Counterparty.Path(), proofs)
	require.NoError(t, err)
	ctx = commitment.WithStore(ctx, store)
	man := node.Manager()
	return ctx, channel.NewHandshaker(man)
}

func (node *Node) CLIState() channel.HandshakeState {
	man := node.Manager()
	return channel.NewHandshaker(man).CLIState(PortName, node.Name, []string{node.Name})
}

func base(cdc *codec.Codec, key sdk.StoreKey) (state.Mapping, state.Mapping) {
	protocol := state.NewMapping(key, cdc, []byte("protocol/"))
	free := state.NewMapping(key, cdc, []byte("free"))
	return protocol, free
}

func (node *Node) Manager() channel.Manager {
	protocol, _ := base(node.Cdc, node.Key)
	_, connman := node.Node.Manager()
	return channel.NewManager(protocol, connman)
}

func (node *Node) OpenInit(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenInit(ctx, PortName, node.Name, node.Channel)
	require.NoError(t, err)
	require.Equal(t, channel.Init, obj.Stage.Get(ctx))
	require.Equal(t, node.Channel, obj.GetChannel(ctx))
	require.False(t, obj.Available.Get(ctx))
}

func (node *Node) OpenTry(t *testing.T, height uint64, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenTry(ctx, proofs, height, PortName, node.Name, node.Channel)
	require.NoError(t, err)
	require.Equal(t, channel.OpenTry, obj.Stage.Get(ctx))
	require.Equal(t, node.Channel, obj.GetChannel(ctx))
	require.False(t, obj.Available.Get(ctx))
	node.SetState(channel.OpenTry)
}

func (node *Node) OpenAck(t *testing.T, height uint64, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenAck(ctx, proofs, height, PortName, node.Name)
	require.NoError(t, err)
	require.Equal(t, channel.Open, obj.Stage.Get(ctx))
	require.Equal(t, node.Channel, obj.GetChannel(ctx))
	require.True(t, obj.Available.Get(ctx))
	node.SetState(channel.Open)
}

func (node *Node) OpenConfirm(t *testing.T, height uint64, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenConfirm(ctx, proofs, height, PortName, node.Name)
	require.NoError(t, err)
	require.Equal(t, channel.Open, obj.Stage.Get(ctx))
	require.Equal(t, node.Channel, obj.GetChannel(ctx))
	require.True(t, obj.Available.Get(ctx))
	node.SetState(channel.CloseTry)
}

func (node *Node) Handshake(t *testing.T) {
	node.Node.Handshake(t)

	// self.OpenInit
	node.OpenInit(t)
	header := node.Commit()

	// counterparty.OpenTry
	node.Counterparty.UpdateClient(t, header)
	cliobj := node.CLIState()
	_, pchan := node.QueryValue(t, cliobj.Channel)
	_, pstate := node.QueryValue(t, cliobj.Stage)
	node.Counterparty.OpenTry(t, uint64(header.Height), pchan, pstate)
	header = node.Counterparty.Commit()

	// self.OpenAck
	node.UpdateClient(t, header)
	cliobj = node.Counterparty.CLIState()
	_, pchan = node.Counterparty.QueryValue(t, cliobj.Channel)
	_, pstate = node.Counterparty.QueryValue(t, cliobj.Stage)
	node.OpenAck(t, uint64(header.Height), pchan, pstate)
	header = node.Commit()

	// counterparty.OpenConfirm
	node.Counterparty.UpdateClient(t, header)
	cliobj = node.CLIState()
	_, pstate = node.QueryValue(t, cliobj.Stage)
	node.Counterparty.OpenConfirm(t, uint64(header.Height), pstate)
}

func (node *Node) Send(t *testing.T, packet channel.Packet) {
	ctx, man := node.Context(), node.Manager()
	obj, err := man.Query(ctx, PortName, node.Name)
	require.NoError(t, err)
	seq := obj.SeqSend.Get(ctx)
	err = man.Send(ctx, node.Name, packet)
	require.NoError(t, err)
	require.Equal(t, seq+1, obj.SeqSend.Get(ctx))
	require.Equal(t, node.Cdc.MustMarshalBinaryBare(packet), obj.PacketCommit(ctx, seq+1))
}

func (node *Node) Receive(t *testing.T, packet channel.Packet, height uint64, proofs ...commitment.Proof) {
	ctx, man := node.Context(), node.Manager()
	obj, err := man.Query(ctx, PortName, node.Name)
	require.NoError(t, err)
	seq := obj.SeqRecv.Get(ctx)
	err = man.Receive(ctx, proofs, height, PortName, node.Name, packet)
	require.NoError(t, err)
	require.Equal(t, seq+1, obj.SeqRecv.Get(ctx))
}
