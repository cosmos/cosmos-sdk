package flag

import (
	"context"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	ip "cosmossdk.io/client/v2/internal/proto"
	"cosmossdk.io/core/coins"
)

type coinType struct{}

type coinValue struct {
	value []*basev1beta1.Coin
}

func (c coinType) NewValue(context.Context, *Builder) Value {
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

	if len(c.value) == 1 {
		return protoreflect.ValueOfMessage(c.value[0].ProtoReflect()), nil
	}

	return protoreflect.ValueOfList(ip.NewGenericList(c.value)), nil
}

func (c *coinValue) String() string {
	if len(c.value) == 1 {
		return c.value[0].String()
	}

	var result string
	for _, coin := range c.value {
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
		c.value = append(c.value, coin)
	}

	return nil
}

func (c *coinValue) Type() string {
	return "cosmos.base.v1beta1.Coin"
}
