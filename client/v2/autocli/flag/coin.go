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
	value []*basev1beta1.Coin
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

	return protoreflect.ValueOfMessage(c.value[0].ProtoReflect()), nil
}

func (c *coinValue) String() string {
	stringCoin, _ := coins.FormatCoins(c.value, nil)
	return stringCoin
}

func (c *coinValue) Set(stringValue string) error {
	coinsStr := strings.Split(stringValue, ",")
	result := make([]*basev1beta1.Coin, len(coinsStr))

	for i, coinStr := range coinsStr {
		coin, err := coins.ParseCoin(coinStr)
		if err != nil {
			return err
		}
		result[i] = coin
	}

	c.value = result
	return nil
}

func (c *coinValue) Type() string {
	return "cosmos.base.v1beta1.Coin"
}
