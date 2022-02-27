package testdata

// DONTCOVER
// nolint

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type Animal interface {
	proto.Message

	Greet() string
}

func (c Cat) Greet() string {
	return fmt.Sprintf("Meow, my name is %s", c.Moniker)
}

func (d Dog) Greet() string {
	return fmt.Sprintf("Roof, my name is %s", d.Name)
}

var _ types.UnpackInterfacesMessage = HasAnimal{} // nolint: exhaustivestruct

func (m HasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal Animal
	return unpacker.UnpackAny(m.Animal, &animal)
}

type HasAnimalI interface {
	TheAnimal() Animal
}

var _ HasAnimalI = &HasAnimal{} // nolint: exhaustivestruct

func (m HasAnimal) TheAnimal() Animal {
	return m.Animal.GetCachedValue().(Animal)
}

type HasHasAnimalI interface {
	TheHasAnimal() HasAnimalI
}

var _ HasHasAnimalI = &HasHasAnimal{} // nolint: exhaustivestruct

func (m HasHasAnimal) TheHasAnimal() HasAnimalI {
	return m.HasAnimal.GetCachedValue().(HasAnimalI)
}

var _ types.UnpackInterfacesMessage = HasHasAnimal{} // nolint: exhaustivestruct

func (m HasHasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal HasAnimalI
	return unpacker.UnpackAny(m.HasAnimal, &animal)
}

type HasHasHasAnimalI interface {
	TheHasHasAnimal() HasHasAnimalI
}

var _ HasHasAnimalI = &HasHasAnimal{} // nolint: exhaustivestruct

func (m HasHasHasAnimal) TheHasHasAnimal() HasHasAnimalI {
	return m.HasHasAnimal.GetCachedValue().(HasHasAnimalI)
}

var _ types.UnpackInterfacesMessage = HasHasHasAnimal{} // nolint: exhaustivestruct

func (m HasHasHasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal HasHasAnimalI
	return unpacker.UnpackAny(m.HasHasAnimal, &animal)
}
