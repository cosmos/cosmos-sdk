package codec

import (
	"testing"

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

	cdc := NewProtoCodec(registry)

	kitty := &testdata.Cat{Moniker: "Kitty"}
	bz, err := MarshalAny(cdc, kitty)
	require.NoError(t, err)

	var animal testdata.Animal

	// empty registry should fail
	err = UnmarshalAny(cdc, &animal, bz)
	require.Error(t, err)

	// wrong type registration should fail
	registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	err = UnmarshalAny(cdc, &animal, bz)
	require.Error(t, err)

	// should pass
	registry = NewTestInterfaceRegistry()
	cdc = NewProtoCodec(registry)
	err = UnmarshalAny(cdc, &animal, bz)
	require.NoError(t, err)
	require.Equal(t, kitty, animal)

	// nil should fail
	registry = NewTestInterfaceRegistry()
	err = UnmarshalAny(cdc, nil, bz)
	require.Error(t, err)
}
