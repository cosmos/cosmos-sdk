package channel

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/tests"
)

type Node struct {
	*connection.Node
	Counterparty *Node
}

func NewNode(self, counter tendermint.MockValidators, cdc *codec.Codec) *Node {
	conn := connection.NewNode(self, counter, cdc)

	res := &Node{
		Node: conn,
	}

	res.Counterparty = &Node{
		Node:         conn.Counterparty,
		Counterparty: res,
	}

	return res
}

func OpenInit(t *testing.T) {
	ctx, man := node.Handshaker(t)

}
