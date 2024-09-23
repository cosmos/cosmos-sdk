package flag

import (
	"context"

	"cosmossdk.io/math"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type decType struct{}

func (a decType) NewValue(_ *context.Context, _ *Builder) Value {
	return &decValue{}
}

func (a decType) DefaultValue() string {
	return "0"
}

type decValue struct {
	value string
}

func (a decValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOf(a.value), nil
}

func (a decValue) String() string {
	return a.value
}

func (a *decValue) Set(s string) error {
	dec, err := math.LegacyNewDecFromStr(s)
	if err != nil {
		return err
	}
	a.value = dec.String()

	return nil
}

func (a decValue) Type() string {
	return "cosmos.Dec"
}
