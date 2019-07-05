package connection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Node struct {
	Name string
	*tendermint.Node
	Counterparty *Node

	Client    string
	Handshake connection.Handshake
	State     connection.State

	Cdc *codec.Codec
}

func NewNode(self, counter tendermint.MockValidators, cdc *codec.Codec) *Node {
	res := &Node{
		Name: "self",                                                           // hard coded, doesnt matter
		Node: tendermint.NewNode(self, tendermint.NewRoot([]byte("protocol"))), // TODO: test with key prefix

		State: connection.Idle,
		Cdc:   cdc,
	}

	res.Counterparty = &Node{
		Name:         "counterparty",
		Node:         tendermint.NewNode(counter, tendermint.NewRoot([]byte("protocol"))),
		Counterparty: res,

		State: connection.Idle,
		Cdc:   cdc,
	}

	res.Handshake = connection.Handshake{
		Counterparty: res.Counterparty.Name,
	}

	res.Counterparty.Handshake = connection.Handshake{
		Counterparty: res.Name,
	}

	return res
}

func (node *Node) CreateClient(t *testing.T) {
	ctx := node.Context()
	climan, _ := node.Manager()
	obj, err := climan.Create(ctx, node.Counterparty.LastStateVerifier().ConsensusState)
	require.NoError(t, err)
	node.Client = obj.ID()
	node.Counterparty.Handshake.CounterpartyClient = obj.ID()
}

func (node *Node) UpdateClient(t *testing.T, header client.Header) {
	ctx := node.Context()
	climan, _ := node.Manager()
	obj, err := climan.Query(ctx, node.Client)
	require.NoError(t, err)
	err = obj.Update(ctx, header)
	require.NoError(t, err)
}

func (node *Node) SetState(state connection.State) {
	node.State = state
	node.Counterparty.State = state
}

func (node *Node) Handshaker(t *testing.T, proofs []commitment.Proof) (sdk.Context, connection.Handshaker) {
	ctx := node.Context()
	store, err := commitment.NewStore(node.Counterparty.Root, proofs)
	require.NoError(t, err)
	ctx = commitment.WithStore(ctx, store)
	_, man := node.Manager()
	return ctx, connection.NewHandshaker(man)
}

func (node *Node) CLIObject() connection.CLIHandshakeObject {
	_, man := node.Manager()
	return connection.NewHandshaker(man).CLIObject(node.Root, node.Name)
}

func base(cdc *codec.Codec, key sdk.StoreKey) (state.Base, state.Base) {
	protocol := state.NewBase(cdc, key, []byte("protocol"))
	free := state.NewBase(cdc, key, []byte("free"))
	return protocol, free
}

func (node *Node) Manager() (client.Manager, connection.Manager) {
	protocol, free := base(node.Cdc, node.Key)
	clientman := client.NewManager(protocol, free, client.IntegerIDGenerator)
	return clientman, connection.NewManager(protocol, clientman)
}

func (node *Node) OpenInit(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenInit(ctx, node.Name, node.Client, node.Handshake, 100) // TODO: test timeout
	require.NoError(t, err)
	require.Equal(t, connection.Init, obj.State(ctx))
	require.Equal(t, node.Handshake, obj.Handshake(ctx))
	require.False(t, obj.Available(ctx))
	node.SetState(connection.Init)
}

func (node *Node) OpenTry(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenTry(ctx, node.Name, node.Client, node.Handshake, 100 /*TODO*/, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, connection.OpenTry, obj.State(ctx))
	require.Equal(t, node.Handshake, obj.Handshake(ctx))
	require.False(t, obj.Available(ctx))
	node.SetState(connection.OpenTry)
}

func (node *Node) OpenAck(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenAck(ctx, node.Name, 100 /*TODO*/, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, connection.Open, obj.State(ctx))
	require.Equal(t, node.Handshake, obj.Handshake(ctx))
	require.True(t, obj.Available(ctx))
	node.SetState(connection.Open)
}

func (node *Node) OpenConfirm(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenConfirm(ctx, node.Name, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, connection.Open, obj.State(ctx))
	require.Equal(t, node.Handshake, obj.Handshake(ctx))
	require.True(t, obj.Available(ctx))
	node.SetState(connection.CloseTry)
}
