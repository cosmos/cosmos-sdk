package testdata

// DONTCOVER
// nolint

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

type Animal interface {
	Greet() string
}

func (c Cat) Greet() string {
	return fmt.Sprintf("Meow, my name is %s", c.Moniker)
}

func (d Dog) Greet() string {
	return fmt.Sprintf("Roof, my name is %s", d.Name)
}

var _ types.UnpackInterfacesMessage = HasAnimal{}

func (m HasAnimal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal Animal
	return unpacker.UnpackAny(m.Animal, &animal)
}
