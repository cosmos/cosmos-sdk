package flag

import (
	"context"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/core/coins"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type coinType struct{}

type coinValue struct {
	value *basev1beta1.Coin
}

func (c coinType) NewValue(_ context.Context, _ *Builder) Value {
	return &coinValue{}
}

func (c coinType) DefaultValue() string {
	stringCoin, _ := coins.FormatCoins([]*basev1beta1.Coin{}, nil)
	return stringCoin
}

func (c *coinValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if c.value == nil {
		return protoreflect.Value{}, nil
	}
	return protoreflect.ValueOfMessage(c.value.ProtoReflect()), nil
}

func (c *coinValue) String() string {
	return c.value.String()
}

func (c *coinValue) Set(stringValue string) error {
	coin, err := coins.ParseCoin(stringValue)
	if err != nil {
		return err
	}
	c.value = coin
	return nil
}

func (c *coinValue) Type() string {
	return "cosmos.base.v1beta1.Coin"
}
