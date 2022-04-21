package cli

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type addressStringFlagType struct{}

func (a addressStringFlagType) AddFlag(_ context.Context, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := &addressStringValue{}
	set.AddFlag(&pflag.Flag{
		Name:  descriptorKebabName(descriptor),
		Usage: descriptorDocs(descriptor),
		Value: val,
	})
	return val
}

type addressStringValue struct {
	value string
}

func (a addressStringValue) Get() protoreflect.Value {
	return protoreflect.ValueOfString(a.value)
}

func (a addressStringValue) String() string {
	return a.value
}

func (a *addressStringValue) Set(s string) error {
	a.value = s
	// TODO handle bech32 validation
	return nil
}

func (a addressStringValue) Type() string {
	return "address"
}
