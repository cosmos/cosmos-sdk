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

	Connection connection.Connection
	State      connection.State

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

	res.Connection = connection.Connection{
		Counterparty: res.Counterparty.Name,
	}

	res.Counterparty.Connection = connection.Connection{
		Counterparty: res.Name,
	}

	return res
}

func (node *Node) CreateClient(t *testing.T) {
	ctx := node.Context()
	climan, _ := node.Manager()
	obj, err := climan.Create(ctx, node.Counterparty.LastStateVerifier().ConsensusState)
	require.NoError(t, err)
	node.Connection.Client = obj.ID()
	node.Counterparty.Connection.CounterpartyClient = obj.ID()
}

func (node *Node) UpdateClient(t *testing.T, header client.Header) {
	ctx := node.Context()
	climan, _ := node.Manager()
	obj, err := climan.Query(ctx, node.Connection.Client)
	require.NoError(t, err)
	err = obj.Update(ctx, header)
	require.NoError(t, err)
}

func (node *Node) SetState(state connection.State) {
	node.State = state
	node.Counterparty.State = state
}

func (node *Node) Object(t *testing.T, proofs []commitment.Proof) (sdk.Context, connection.Object) {
	ctx := node.Context()
	store, err := commitment.NewStore(node.Counterparty.Root, proofs)
	require.NoError(t, err)
	ctx = commitment.WithStore(ctx, store)
	_, man := node.Manager()
	switch node.State {
	case connection.Idle, connection.Init:
		obj, err := man.Create(ctx, node.Name, node.Connection)
		require.NoError(t, err)
		return ctx, obj
	default:
		obj, err := man.Query(ctx, node.Name)
		require.NoError(t, err)
		require.Equal(t, node.Connection, obj.Connection(ctx))
		return ctx, obj
	}
}

func (node *Node) CLIObject() connection.CLIObject {
	_, man := node.Manager()
	return man.CLIObject(node.Root, node.Name)
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

func (node *Node) Advance(t *testing.T, proofs ...commitment.Proof) {
	switch node.State {
	// TODO: use different enum type for node.State
	case connection.Idle: // self: Idle -> Init
		ctx, obj := node.Object(t, proofs)
		require.Equal(t, connection.Idle, obj.State(ctx))
		err := obj.OpenInit(ctx, 100) // TODO: test timeout
		require.NoError(t, err)
		require.Equal(t, connection.Init, obj.State(ctx))
		require.Equal(t, node.Connection, obj.Connection(ctx))
		node.SetState(connection.Init)
	case connection.Init: // counter: Idle -> OpenTry
		ctx, obj := node.Counterparty.Object(t, proofs)
		require.Equal(t, connection.Idle, obj.State(ctx))
		err := obj.OpenTry(ctx, 0 /*TODO*/, 100 /*TODO*/, 100 /*TODO*/)
		require.NoError(t, err)
		require.Equal(t, connection.OpenTry, obj.State(ctx))
		require.Equal(t, node.Counterparty.Connection, obj.Connection(ctx))
		node.SetState(connection.OpenTry)
	case connection.OpenTry: // self: Init -> Open
		ctx, obj := node.Object(t, proofs)
		require.Equal(t, connection.Init, obj.State(ctx))
		err := obj.OpenAck(ctx, 0 /*TODO*/, 100 /*TODO*/, 100 /*TODO*/)
		require.NoError(t, err)
		require.Equal(t, connection.Open, obj.State(ctx))
		require.Equal(t, node.Connection, obj.Connection(ctx))
		node.SetState(connection.Open)
	case connection.Open: // counter: OpenTry -> Open
		ctx, obj := node.Counterparty.Object(t, proofs)
		require.Equal(t, connection.OpenTry, obj.State(ctx))
		err := obj.OpenConfirm(ctx, 100 /*TODO*/)
		require.NoError(t, err)
		require.Equal(t, connection.Open, obj.State(ctx))
		require.Equal(t, node.Counterparty.Connection, obj.Connection(ctx))
		node.SetState(connection.CloseTry)
		// case connection.CloseTry // self: Open -> CloseTry
	default:
		return
	}
}
