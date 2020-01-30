package testdata

// DONTCOVER
// nolint

import (
	"fmt"
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
