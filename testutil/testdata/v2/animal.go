package testdata

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
)

var _ Animal = &Snake{}

type Animal interface {
	proto.Message

	Greet() string
}

func (x Snake) Greet() string {
	return fmt.Sprintf("*violent snake noises*, my name is %s", x.Name)
}
