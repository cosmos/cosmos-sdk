package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
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

func TestPackUnpack(t *testing.T) {
	registry := NewTestInterfaceRegistry()

	spot := &testdata.Dog{Name: "Spot"}
	any := types.Any{}
	err := any.Pack(spot)
	require.NoError(t, err)

	require.Equal(t, spot, any.GetCachedValue())

	// without cache
	any.ClearCachedValue()
	var animal testdata.Animal
	err = registry.UnpackAny(&any, &animal)
	require.NoError(t, err)
	require.Equal(t, spot, animal)

	// with cache
	err = any.Pack(spot)
	require.Equal(t, spot, any.GetCachedValue())
	require.NoError(t, err)
	err = registry.UnpackAny(&any, &animal)
	require.NoError(t, err)
	require.Equal(t, spot, animal)
}

func TestMarshalAny(t *testing.T) {
	registry := types.NewInterfaceRegistry()

	kitty := &testdata.Cat{Moniker: "Kitty"}
	bz, err := types.MarshalAny(kitty)
	require.NoError(t, err)

	var animal testdata.Animal

	// empty registry should fail
	err = types.UnmarshalAny(registry, &animal, bz)
	require.Error(t, err)

	// wrong type registration should fail
	registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	err = types.UnmarshalAny(registry, &animal, bz)
	require.Error(t, err)

	// should pass
	registry = NewTestInterfaceRegistry()
	err = types.UnmarshalAny(registry, &animal, bz)
	require.NoError(t, err)
	require.Equal(t, kitty, animal)

	// nil should fail
	registry = NewTestInterfaceRegistry()
	err = types.UnmarshalAny(registry, nil, bz)
}

type TestI interface {
	DoSomething()
}

func TestRegister(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*testdata.Animal)(nil))
	registry.RegisterInterface("TestI", (*TestI)(nil))
	require.NotPanics(t, func() {
		registry.RegisterImplementations((*testdata.Animal)(nil), &testdata.Dog{})
	})
	require.Panics(t, func() {
		registry.RegisterImplementations((*TestI)(nil), &testdata.Dog{})
	})
	require.Panics(t, func() {
		registry.RegisterImplementations((*TestI)(nil), nil)
	})
}

func TestUnpackInterfaces(t *testing.T) {
	registry := NewTestInterfaceRegistry()

	spot := &testdata.Dog{Name: "Spot"}
	any, err := types.NewAnyWithValue(spot)
	require.NoError(t, err)

	hasAny := testdata.HasAnimal{
		Animal: any,
		X:      1,
	}
	bz, err := hasAny.Marshal()
	require.NoError(t, err)

	var hasAny2 testdata.HasAnimal
	err = hasAny2.Unmarshal(bz)
	require.NoError(t, err)

	err = types.UnpackInterfaces(hasAny2, registry)
	require.NoError(t, err)

	require.Equal(t, spot, hasAny2.Animal.GetCachedValue())
}
