package cli

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type addressStringFlagType struct{}

func (a addressStringFlagType) AddFlag(_ context.Context, builder *Builder, set *pflag.FlagSet, descriptor protoreflect.FieldDescriptor) FlagValue {
	val := &addressStringFlagValue{}
	set.AddFlag(&pflag.Flag{
		Name:  descriptorKebabName(descriptor),
		Usage: descriptorDocs(descriptor),
		Value: val,
	})
	return val
}

type addressStringFlagValue struct {
	value string
}

func (a addressStringFlagValue) Get() protoreflect.Value {
	return protoreflect.ValueOfString(a.value)
}

func (a addressStringFlagValue) String() string {
	return a.value
}

func (a *addressStringFlagValue) Set(s string) error {
	a.value = s
	// TODO handle bech32 validation
	return nil
}

func (a addressStringFlagValue) Type() string {
	return "bech32 account address key name"
}
