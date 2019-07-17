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
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

type Node struct {
	Name string
	*tendermint.Node
	Counterparty *Node

	CounterpartyClient string
	Connection         connection.Connection
	State              connection.State

	Cdc *codec.Codec
}

func NewNode(self, counter tendermint.MockValidators, cdc *codec.Codec) *Node {
	res := &Node{
		Name: "self",                                                                                          // hard coded, doesnt matter
		Node: tendermint.NewNode(self, merkle.NewPath([][]byte{[]byte("teststoreself")}, []byte("protocol"))), // TODO: test with key prefix

		State: connection.Idle,
		Cdc:   cdc,
	}

	res.Counterparty = &Node{
		Name:         "counterparty",
		Node:         tendermint.NewNode(counter, merkle.NewPath([][]byte{[]byte("teststorecounterparty")}, []byte("protocol"))),
		Counterparty: res,

		State: connection.Idle,
		Cdc:   cdc,
	}

	res.Connection = connection.Connection{
		Counterparty: res.Counterparty.Name,
		Path:         res.Counterparty.Path,
	}

	res.Counterparty.Connection = connection.Connection{
		Counterparty: res.Name,
		Path:         res.Path,
	}

	return res
}

func (node *Node) CreateClient(t *testing.T) {
	ctx := node.Context()
	climan, _ := node.Manager()
	obj, err := climan.Create(ctx, node.Name, node.Counterparty.LastStateVerifier().ConsensusState)
	require.NoError(t, err)
	node.Connection.Client = obj.ID()
	node.Counterparty.CounterpartyClient = obj.ID()
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

func (node *Node) Handshaker(t *testing.T, proofs []commitment.Proof) (sdk.Context, connection.Handshaker) {
	ctx := node.Context()
	_, man := node.Manager()
	return ctx, connection.NewHandshaker(man)
}

func (node *Node) CLIObject() connection.CLIHandshakeObject {
	_, man := node.Manager()
	return connection.NewHandshaker(man).CLIObject(node.Path, node.Name, node.Name)
}

func base(cdc *codec.Codec, key sdk.StoreKey) state.Base {
	protocol := state.NewBase(cdc, key, []byte("protocol"))
	return protocol
}

func (node *Node) Manager() (client.Manager, connection.Manager) {
	protocol := base(node.Cdc, node.Key)
	clientman := client.NewManager(protocol)
	return clientman, connection.NewManager(protocol, clientman)
}

func (node *Node) OpenInit(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenInit(ctx, node.Name, node.Connection, node.CounterpartyClient, 100) // TODO: test timeout
	require.NoError(t, err)
	require.Equal(t, connection.Init, obj.State(ctx))
	require.Equal(t, node.Connection, obj.Connection(ctx))
	require.Equal(t, node.CounterpartyClient, obj.CounterpartyClient(ctx))
	require.False(t, obj.Available(ctx))
	node.SetState(connection.Init)
}

func (node *Node) OpenTry(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenTry(ctx, proofs, node.Name, node.Connection, node.CounterpartyClient, 100 /*TODO*/, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, connection.OpenTry, obj.State(ctx))
	require.Equal(t, node.Connection, obj.Connection(ctx))
	require.Equal(t, node.CounterpartyClient, obj.CounterpartyClient(ctx))
	require.False(t, obj.Available(ctx))
	node.SetState(connection.OpenTry)
}

func (node *Node) OpenAck(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenAck(ctx, proofs, node.Name, 100 /*TODO*/, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, connection.Open, obj.State(ctx))
	require.Equal(t, node.Connection, obj.Connection(ctx))
	require.True(t, obj.Available(ctx))
	node.SetState(connection.Open)
}

func (node *Node) OpenConfirm(t *testing.T, proofs ...commitment.Proof) {
	ctx, man := node.Handshaker(t, proofs)
	obj, err := man.OpenConfirm(ctx, proofs, node.Name, 100 /*TODO*/)
	require.NoError(t, err)
	require.Equal(t, connection.Open, obj.State(ctx))
	require.Equal(t, node.Connection, obj.Connection(ctx))
	require.True(t, obj.Available(ctx))
	node.SetState(connection.CloseTry)
}
