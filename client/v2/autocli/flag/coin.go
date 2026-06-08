package flag

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"cosmossdk.io/client/v2/internal/coins"
	coinscore "cosmossdk.io/core/coins"
)

type coinType struct{}

type coinValue struct {
	value *coinscore.Coin
}

func (c coinType) NewValue(*context.Context, *Builder) Value {
	return &coinValue{}
}

func (c coinType) DefaultValue() string {
	return "zero"
}

func (c *coinValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if c.value == nil {
		return protoreflect.Value{}, nil
	}
	// Build a dynamic proto message for cosmos.base.v1beta1.Coin using the
	// global type registry (populated by gogoproto's init()).
	mt, err := protoregistry.GlobalTypes.FindMessageByName("cosmos.base.v1beta1.Coin")
	if err != nil {
		return protoreflect.Value{}, err
	}
	dynCoin := dynamicpb.NewMessage(mt.Descriptor())
	dynCoin.Set(mt.Descriptor().Fields().ByName("denom"), protoreflect.ValueOfString(c.value.Denom))
	dynCoin.Set(mt.Descriptor().Fields().ByName("amount"), protoreflect.ValueOfString(c.value.Amount))
	return protoreflect.ValueOfMessage(dynCoin), nil
}

func (c *coinValue) String() string {
	if c.value == nil {
		return ""
	}
	return c.value.Amount + c.value.Denom
}

func (c *coinValue) Set(stringValue string) error {
	if strings.Contains(stringValue, ",") {
		return errors.New("coin flag must be a single coin, specific multiple coins with multiple flags or spaces")
	}
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
