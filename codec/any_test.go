package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
	bz, err := cdc.MarshalInterface(kitty)
	require.NoError(t, err)

	var animal testdata.Animal

	// empty registry should fail
	err = cdc.UnmarshalInterface(bz, &animal)
	require.Error(t, err)

	// wrong type registration should fail
	registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	err = cdc.UnmarshalInterface(bz, &animal)
	require.Error(t, err)

	// should pass
	registry = NewTestInterfaceRegistry()
	cdc = codec.NewProtoCodec(registry)
	err = cdc.UnmarshalInterface(bz, &animal)
	require.NoError(t, err)
	require.Equal(t, kitty, animal)

	// nil should fail
	registry = NewTestInterfaceRegistry()
	err = cdc.UnmarshalInterface(bz, nil)
	require.Error(t, err)
}
