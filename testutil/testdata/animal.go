package testdata

// DONTCOVER
// nolint

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type Animal interface {
	proto.Message

	Greet() string
}

type Cartoon interface {
	proto.Message

	Identify() string
}

func (c *Cat) Greet() string {
	return fmt.Sprintf("Meow, my name is %s", c.Moniker)
}

func (c *Bird) Identify() string {
	return "This is Tweety."
}

func (d Dog) Greet() string {
	return fmt.Sprintf("Roof, my name is %s", d.Name)
}

var _ types.UnpackInterfacesMessage = HasAnimal{}

func (m HasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal Animal
	return unpacker.UnpackAny(m.Animal, &animal)
}

type IHasAnimal interface {
	TheAnimal() Animal
}

var _ IHasAnimal = &HasAnimal{}

func (m HasAnimal) TheAnimal() Animal {
	return m.Animal.GetCachedValue().(Animal)
}

type IHasHasAnimal interface {
	TheHasAnimal() IHasAnimal
}

var _ IHasHasAnimal = &HasHasAnimal{}

func (m HasHasAnimal) TheHasAnimal() IHasAnimal {
	return m.HasAnimal.GetCachedValue().(IHasAnimal)
}

var _ types.UnpackInterfacesMessage = HasHasAnimal{}

func (m HasHasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal IHasAnimal
	return unpacker.UnpackAny(m.HasAnimal, &animal)
}

type IHasHasHasAnimal interface {
	TheHasHasAnimal() IHasHasAnimal
}

var _ IHasHasAnimal = &HasHasAnimal{}

func (m HasHasHasAnimal) TheHasHasAnimal() IHasHasAnimal {
	return m.HasHasAnimal.GetCachedValue().(IHasHasAnimal)
}

var _ types.UnpackInterfacesMessage = HasHasHasAnimal{}

func (m HasHasHasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal IHasHasAnimal
	return unpacker.UnpackAny(m.HasHasAnimal, &animal)
}
