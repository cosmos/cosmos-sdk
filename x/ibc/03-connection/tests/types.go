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
)

type Node struct {
	Name string
	*tendermint.Node
	Counterparty *CounterpartyNode

	State connection.State

	Cdc *codec.Codec
}

func NewNode(self, counter tendermint.MockValidators, cdc *codec.Codec) *Node {
	res := &Node{
		Name: "self", // hard coded, doesnt matter
		Node: tendermint.NewNode(self),

		State: connection.Idle,
		Cdc:   cdc,
	}

	res.Counterparty = &CounterpartyNode{
		Name: "counterparty",
		Node: tendermint.NewNode(counter),
		Cdc:  cdc,

		SelfName: &res.Name,
		State:    &res.State,
	}

	return res
}

func (node *Node) Object(t *testing.T) (sdk.Context, connection.Object) {
	ctx, man, conn := node.Context(), node.Manager(), node.Connection()
	switch node.State {
	case connection.Idle:
		obj, err := man.Create(ctx, node.Name, conn)
		require.NoError(t, err)
		return ctx, obj
	default:
		obj, err := man.Query(ctx, node.Name)
		require.NoError(t, err)
		require.Equal(t, conn, obj.Connection(ctx))
		return ctx, obj
	}
}

func (node *Node) Manager() connection.Manager {
	base := state.NewBase(node.Cdc, node.Key)
	protocol := base.Prefix([]byte("protocol/"))
	free := base.Prefix([]byte("free/"))
	clientman := client.NewManager(protocol, free, client.IntegerIDGenerator)
	return connection.NewManager(base, clientman)
}

func (node *Node) Connection() connection.Connection {
	return connection.Connection{
		Counterparty:       node.Counterparty.Name,
		Client:             node.Name + "client",
		CounterpartyClient: node.Counterparty.Name + "client",
	}
}

type CounterpartyNode struct {
	Name string
	*tendermint.Node

	Cdc *codec.Codec

	// pointer to self
	// TODO: improve
	State    *connection.State
	SelfName *string
}

func (node *CounterpartyNode) Object(t *testing.T) (sdk.Context, connection.Object) {
	ctx, man, conn := node.Context(), node.Manager(), node.Connection()
	switch *node.State {
	case connection.Idle:
		obj, err := man.Create(ctx, node.Name, conn)
		require.NoError(t, err)
		return ctx, obj
	default:
		obj, err := man.Query(ctx, node.Name)
		require.NoError(t, err)
		require.Equal(t, conn, obj.Connection(ctx))
		return ctx, obj
	}
}
func (node *CounterpartyNode) Connection() connection.Connection {
	return connection.Connection{
		Counterparty:       *node.SelfName,
		Client:             node.Name + "client",
		CounterpartyClient: *node.SelfName + "client",
	}
}

func (node *CounterpartyNode) Manager() connection.Manager {
	base := state.NewBase(node.Cdc, node.Key)
	protocol := base.Prefix([]byte("protocol/"))
	free := base.Prefix([]byte("free/"))
	clientman := client.NewManager(protocol, free, client.IntegerIDGenerator)
	return connection.NewManager(base, clientman)
}

func (node *Node) Advance(t *testing.T) {
	switch node.State {
	case connection.Idle: // self: Idle -> Init
		ctx, obj := node.Object(t)
		require.Equal(t, connection.Idle, obj.State(ctx))
		err := obj.OpenInit(ctx, 100) // TODO: test timeout
		require.NoError(t, err)
		require.Equal(t, connection.Init, obj.State(ctx))
		require.Equal(t, node.Connection(), obj.Connection(ctx))
		node.State = connection.Init
	case connection.Init: // counter: Idle -> OpenTry
		ctx, obj := node.Counterparty.Object(t)
		require.Equal(t, connection.Idle, obj.State(ctx))
		err := obj.OpenTry(ctx, 0 /*TODO*/, 100 /*TODO*/, 100 /*TODO*/)
		require.NoError(t, err)
		require.Equal(t, connection.OpenTry, obj.State(ctx))
		require.Equal(t, node.Counterparty.Connection(), obj.Connection(ctx))
		node.State = connection.OpenTry
	case connection.OpenTry: // self: Init -> Open
		ctx, obj := node.Object(t)
		require.Equal(t, connection.Init, obj.State(ctx))
		err := obj.OpenAck(ctx, 0 /*TODO*/, 100 /*TODO*/, 100 /*TODO*/)
		require.NoError(t, err)
		require.Equal(t, connection.Open, obj.State(ctx))
		require.Equal(t, node.Connection(), obj.Connection(ctx))
		node.State = connection.Open
	case connection.Open: // counter: OpenTry -> Open
		ctx, obj := node.Counterparty.Object(t)
		require.Equal(t, connection.OpenTry, obj.State(ctx))
		err := obj.OpenConfirm(ctx, 100 /*TODO*/)
		require.NoError(t, err)
		require.Equal(t, connection.Open, obj.State(ctx))
		require.Equal(t, node.Counterparty.Connection(), obj.Connection(ctx))
		node.State = connection.CloseTry
	// case connection.CloseTry // self: Open -> CloseTry
	default:
		return
	}
}
