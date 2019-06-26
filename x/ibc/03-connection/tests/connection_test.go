package connection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint/tests"
)

func registerCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
}

func TestHandshake(t *testing.T) {
	cdc := codec.New()

	node := NewNode(tendermint.NewMockValidators(10), tendermint.NewMockValidators(10))
}
