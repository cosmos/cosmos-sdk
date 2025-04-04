package flag

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type addressStringType struct{}

func (a addressStringType) NewValue(_ context.Context, _ *Builder) Value {
	return &addressValue{}
}

func (a addressStringType) DefaultValue() string {
	return ""
}

type addressValue struct {
	value string
}

func (a addressValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfString(a.value), nil
}

func (a addressValue) String() string {
	return a.value
}

func (a *addressValue) Set(s string) error {
	a.value = s
	// TODO handle bech32 validation
	return nil
}

func (a addressValue) Type() string {
	return "bech32 account address key name"
}
