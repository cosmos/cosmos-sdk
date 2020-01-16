package testdata

// DONTCOVER
// nolint

import (
	"fmt"
)

type Animal interface {
	Greet() string
}

// Cat is not defined a Protobuf message.
type Cat struct {
	Name string
}

func (c Cat) Greet() string {
	return fmt.Sprintf("Meow, my name is %s", c.Name)
}

func (d Dog) Greet() string {
	return fmt.Sprintf("Roof, my name is %s", d.Name)
}
