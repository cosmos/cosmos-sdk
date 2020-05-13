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
	registry.RegisterImplementations(
		(*testdata.HasAnimalI)(nil),
		&testdata.HasAnimal{},
	)
	registry.RegisterImplementations(
		(*testdata.HasHasAnimalI)(nil),
		&testdata.HasHasAnimal{},
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
	require.Panics(t, func() {
		registry.RegisterInterface("not_an_interface", (*testdata.Dog)(nil))
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

func TestNested(t *testing.T) {
	registry := NewTestInterfaceRegistry()

	spot := &testdata.Dog{Name: "Spot"}
	any, err := types.NewAnyWithValue(spot)
	require.NoError(t, err)

	ha := &testdata.HasAnimal{Animal: any}
	any2, err := types.NewAnyWithValue(ha)
	require.NoError(t, err)

	hha := &testdata.HasHasAnimal{HasAnimal: any2}
	any3, err := types.NewAnyWithValue(hha)
	require.NoError(t, err)

	hhha := testdata.HasHasHasAnimal{HasHasAnimal: any3}

	// marshal
	bz, err := hhha.Marshal()
	require.NoError(t, err)

	// unmarshal
	var hhha2 testdata.HasHasHasAnimal
	err = hhha2.Unmarshal(bz)
	require.NoError(t, err)
	err = types.UnpackInterfaces(hhha2, registry)
	require.NoError(t, err)

	require.Equal(t, spot, hhha2.TheHasHasAnimal().TheHasAnimal().TheAnimal())
}
