package flag

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/address"
)

type addressStringType struct{}

func (a addressStringType) NewValue(ctx context.Context, b *Builder) Value {
	return &addressValue{addressCodec: b.AddressCodec}
}

func (a addressStringType) DefaultValue() string {
	return ""
}

type validatorAddressStringType struct{}

func (a validatorAddressStringType) NewValue(ctx context.Context, b *Builder) Value {
	return &addressValue{addressCodec: b.ValidatorAddressCodec}
}

func (a validatorAddressStringType) DefaultValue() string {
	return ""
}

type addressValue struct {
	value        string
	addressCodec address.Codec
}

func (a addressValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfString(a.value), nil
}

func (a addressValue) String() string {
	return a.value
}

// Set implements the flag.Value interface for addressValue it only supports bech32 addresses.
func (a *addressValue) Set(s string) error {
	_, err := a.addressCodec.StringToBytes(s)
	if err != nil {
		return fmt.Errorf("invalid bech32 account address: %w", err)
	}

	a.value = s

	return nil
}

func (a addressValue) Type() string {
	return "bech32 account address key name"
}
