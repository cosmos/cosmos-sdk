package testdata

import (
	"github.com/cosmos/gogoproto/types/any/test"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	tx "github.com/cosmos/cosmos-sdk/types/tx"
)

func NewTestInterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	return registry
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &TestMsg{})

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
	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&Cat{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func NewTestAmino() *amino.Codec {
	cdc := amino.NewCodec()
	cdc.RegisterInterface((*test.Animal)(nil), nil)
	cdc.RegisterConcrete(&test.Dog{}, "testpb/Dog", nil)
	cdc.RegisterConcrete(&test.Cat{}, "testpb/Cat", nil)

	cdc.RegisterInterface((*test.HasAnimalI)(nil), nil)
	cdc.RegisterConcrete(&test.HasAnimal{}, "testpb/HasAnimal", nil)

	cdc.RegisterInterface((*test.HasHasAnimalI)(nil), nil)
	cdc.RegisterConcrete(&test.HasHasAnimal{}, "testpb/HasHasAnimal", nil)

	return cdc
}
