package connection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	tmclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

func registerCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
	tmclient.RegisterCodec(cdc)
	commitment.RegisterCodec(cdc)
	merkle.RegisterCodec(cdc)
}

func TestHandshake(t *testing.T) {
	cdc := codec.New()
	registerCodec(cdc)

	node := NewNode(tendermint.NewMockValidators(10), tendermint.NewMockValidators(10), cdc)
	node.Advance(t)
	node.Advance(t)
	node.Advance(t)
	node.Advance(t)
}
