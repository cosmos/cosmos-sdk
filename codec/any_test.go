package codec_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

func NewTestInterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*testdata.Animal)(nil))
	registry.RegisterImplementations(
		(*testdata.Animal)(nil),
		&testdata.Dog{},
		&testdata.Cat{},
	)
	return registry
}

func TestMarshalAny(t *testing.T) {
	registry := types.NewInterfaceRegistry()

	cdc := codec.NewProtoCodec(registry)

	kitty := &testdata.Cat{Moniker: "Kitty"}
	bz, err := codec.MarshalAny(cdc, kitty)
	require.NoError(t, err)

	var animal testdata.Animal

	// empty registry should fail
	err = codec.UnmarshalAny(cdc, &animal, bz)
	require.Error(t, err)

	// wrong type registration should fail
	registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	err = codec.UnmarshalAny(cdc, &animal, bz)
	require.Error(t, err)

	// should pass
	registry = NewTestInterfaceRegistry()
	cdc = codec.NewProtoCodec(registry)
	err = codec.UnmarshalAny(cdc, &animal, bz)
	require.NoError(t, err)
	require.Equal(t, kitty, animal)

	// nil should fail
	registry = NewTestInterfaceRegistry()
	err = codec.UnmarshalAny(cdc, nil, bz)
	require.Error(t, err)
}
