package testdata

import (
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

func NewTestInterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*Animal)(nil))
	registry.RegisterImplementations(
		(*Animal)(nil),
		&Dog{},
		&Cat{},
	)
	registry.RegisterImplementations(
		(*HasAnimalI)(nil),
		&HasAnimal{},
	)
	registry.RegisterImplementations(
		(*HasHasAnimalI)(nil),
		&HasHasAnimal{},
	)
	return registry
}

func NewTestAmino() *amino.Codec {
	cdc := amino.NewCodec()
	cdc.RegisterInterface((*Animal)(nil), nil)
	cdc.RegisterConcrete(&Dog{}, "testdata/Dog", nil)
	cdc.RegisterConcrete(&Cat{}, "testdata/Cat", nil)

	cdc.RegisterInterface((*HasAnimalI)(nil), nil)
	cdc.RegisterConcrete(&HasAnimal{}, "testdata/HasAnimal", nil)

	cdc.RegisterInterface((*HasHasAnimalI)(nil), nil)
	cdc.RegisterConcrete(&HasHasAnimal{}, "testdata/HasHasAnimal", nil)

	return cdc
}
