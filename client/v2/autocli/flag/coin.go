package flag

import (
	"context"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/core/coins"
)

type coinType struct{}

type coinValue struct {
	values []*basev1beta1.Coin
}

func (c coinType) NewValue(context.Context, *Builder) Value {
	return &coinValue{}
}

func (c coinType) DefaultValue() string {
	stringCoin, _ := coins.FormatCoins([]*basev1beta1.Coin{}, nil)
	return stringCoin
}

func (c *coinValue) Get(mutable protoreflect.Value) (protoreflect.Value, error) {
	if c.values == nil {
		return protoreflect.Value{}, nil
	}

	list := mutable.List()
	for _, value := range c.values {
		list.Append(protoreflect.ValueOfMessage(value.ProtoReflect()))
	}

	return mutable, nil
}

func (c *coinValue) String() string {
	if len(c.values) == 1 {
		return c.values[0].String()
	}

	var result string
	for _, coin := range c.values {
		result += coin.String() + ","
	}

	return result
}

func (c *coinValue) Set(stringValue string) error {
	result := strings.Split(stringValue, ",")
	for _, coin := range result {
		coin, err := coins.ParseCoin(coin)
		if err != nil {
			return err
		}
		c.values = append(c.values, coin)
	}

	return nil
}

func (c *coinValue) Type() string {
	return "cosmos.base.v1beta1.Coin"
}
