package channel

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Node struct {
	*connection.Node
	Counterparty *Node

	Channel channel.Channel

	Cdc *codec.Codec
}

func NewNode(self, counter tendermint.MockValidators, cdc *codec.Codec) *Node {
	res := &Node{
		Node: connection.NewNode(self, counter, cdc), // TODO: test with key prefix

		Cdc: cdc,
	}

	res.Counterparty = &Node{
		Node:         res.Node.Counterparty,
		Counterparty: res,

		Cdc: cdc,
	}

	res.Channel = channel.Channel{
		Port:             "", // TODO
		Counterparty:     res.Counterparty.Name,
		CounterpartyPort: "", //TODO
	}

	res.Counterparty.Channel = channel.Channel{
		Port:             "", // TODO
		Counterparty:     res.Name,
		CounterpartyPort: "", // TODO
	}

	return res
}

func (node *Node) Handshaker(t *testing.T, proofs []commitment.Proof) (sdk.Context, channel.Handshaker) {
	ctx := node.Context()
	store, err := commitment.NewStore(node.Counterparty.Root, proofs)
	require.NoError(t, err)
	ctx = commitment.WithStore(ctx, store)
	man := node.Manager()
	return ctx, channel.NewHandshaker(man)
}

func (node *Node) CLIObject() channel.CLIHandshakeObject {
	man := node.Manager()
	return channel.NewHandshaker(man).CLIObject(node.Root, node.Name, node.Name)
}

func base(cdc *codec.Codec, key sdk.StoreKey) (state.Base, state.Base) {
	protocol := state.NewBase(cdc, key, []byte("protocol"))
	free := state.NewBase(cdc, key, []byte("free"))
	return protocol, free
}

func (node *Node) Manager() channel.Manager {
	protocol, _ := base(node.Cdc, node.Key)
	_, connman := node.Node.Manager()
	return channel.NewManager(protocol, connman)
}

func (node *Node) OpenInit(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenInit(ctx, node.Name, node.Name, node.Channel, 100) // TODO: test timeout
	require.NoError(t, err)
	require.Equal(t, channel.Init, obj.State(ctx))
	require.Equal(t, node.Channel, obj.Channel(ctx))
	require.False(t, obj.Available(ctx))
}

func (node *Node) OpenTry(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenTry(ctx, node.Name, node.Name, node.Channel, 100 /*TODO*/, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, channel.OpenTry, obj.State(ctx))
	require.Equal(t, node.Channel, obj.Channel(ctx))
	require.False(t, obj.Available(ctx))
	node.SetState(channel.OpenTry)
}

func (node *Node) OpenAck(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenAck(ctx, node.Name, node.Name, 100 /*TODO*/, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, channel.Open, obj.State(ctx))
	require.Equal(t, node.Channel, obj.Channel(ctx))
	require.True(t, obj.Available(ctx))
	node.SetState(channel.Open)
}

func (node *Node) OpenConfirm(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenConfirm(ctx, node.Name, node.Name, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, channel.Open, obj.State(ctx))
	require.Equal(t, node.Channel, obj.Channel(ctx))
	require.True(t, obj.Available(ctx))
	node.SetState(channel.CloseTry)
}

func (node *Node) Handshake(t *testing.T) {
	node.Node.Handshake(t)

	// self.OpenInit
	node.OpenInit(t)
	header := node.Commit()

	// counterparty.OpenTry
	node.Counterparty.UpdateClient(t, header)
	cliobj := node.CLIObject()
	_, pchan := node.Query(t, cliobj.ChannelKey)
	_, pstate := node.Query(t, cliobj.StateKey)
	_, ptimeout := node.Query(t, cliobj.TimeoutKey)
	node.Counterparty.OpenTry(t, pchan, pstate, ptimeout)
	header = node.Counterparty.Commit()

	// self.OpenAck
	node.UpdateClient(t, header)
	cliobj = node.Counterparty.CLIObject()
	_, pchan = node.Counterparty.Query(t, cliobj.ChannelKey)
	_, pstate = node.Counterparty.Query(t, cliobj.StateKey)
	_, ptimeout = node.Counterparty.Query(t, cliobj.TimeoutKey)
	node.OpenAck(t, pchan, pstate, ptimeout)
	header = node.Commit()

	// counterparty.OpenConfirm
	node.Counterparty.UpdateClient(t, header)
	cliobj = node.CLIObject()
	_, pstate = node.Query(t, cliobj.StateKey)
	_, ptimeout = node.Query(t, cliobj.TimeoutKey)
	node.Counterparty.OpenConfirm(t, pstate, ptimeout)
}

func (node *Node) Send(t *testing.T, packet channel.Packet) {
	ctx, man := node.Context(), node.Manager()
	obj, err := man.Query(ctx, node.Name, node.Name)
	require.NoError(t, err)
	seq := obj.SeqSend(ctx)
	err = obj.Send(ctx, packet)
	require.NoError(t, err)
	require.Equal(t, seq+1, obj.SeqSend(ctx))
	require.Equal(t, packet, obj.Packet(ctx, seq+1))
}

func (node *Node) Receive(t *testing.T, packet channel.Packet) {
	ctx, man := node.Context(), node.Manager()
	obj, err := man.Query(ctx, node.Name, node.Name)
	require.NoError(t, err)
	seq := obj.SeqRecv(ctx)
	err = obj.Receive(ctx, packet)
	require.NoError(t, err)
	require.Equal(t, seq+1, obj.SeqRecv(ctx))
}
