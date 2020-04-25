package types

import (
	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestI interface{}

var _ TestI = &testdata.Dog{}

func TestAny_Pack(t *testing.T) {
	ctx := NewInterfaceContext()
	ctx.RegisterInterface("cosmos_sdk.test.TestI", (*TestI)(nil))
	ctx.RegisterImplementation((*TestI)(nil), &testdata.Dog{})
	ctx.RegisterImplementation((*TestI)(nil), &testdata.Cat{})

	spot := &testdata.Dog{Name: "Spot"}
	any := Any{}
	err := any.Pack(spot)
	require.NoError(t, err)

	var test TestI
	err = ctx.UnpackAny(&any, &test)
	require.NoError(t, err)
	require.Equal(t, spot, test)
}
