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

type decCoinType struct{}

type decCoinValue struct {
	value *coinscore.DecCoin
}

func (c decCoinType) NewValue(*context.Context, *Builder) Value {
	return &decCoinValue{}
}

func (c decCoinType) DefaultValue() string {
	return "zero"
}

func (c *decCoinValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if c.value == nil {
		return protoreflect.Value{}, nil
	}
	// Build a dynamic proto message for cosmos.base.v1beta1.DecCoin using the
	// global type registry (populated by gogoproto's init()).
	mt, err := protoregistry.GlobalTypes.FindMessageByName("cosmos.base.v1beta1.DecCoin")
	if err != nil {
		return protoreflect.Value{}, err
	}
	dynCoin := dynamicpb.NewMessage(mt.Descriptor())
	dynCoin.Set(mt.Descriptor().Fields().ByName("denom"), protoreflect.ValueOfString(c.value.Denom))
	dynCoin.Set(mt.Descriptor().Fields().ByName("amount"), protoreflect.ValueOfString(c.value.Amount))
	return protoreflect.ValueOfMessage(dynCoin), nil
}

func (c *decCoinValue) String() string {
	if c.value == nil {
		return ""
	}
	return c.value.Amount + c.value.Denom
}

func (c *decCoinValue) Set(stringValue string) error {
	if strings.Contains(stringValue, ",") {
		return errors.New("coin flag must be a single coin, specific multiple coins with multiple flags or spaces")
	}
	coin, err := coins.ParseDecCoin(stringValue)
	if err != nil {
		return err
	}
	c.value = coin
	return nil
}

func (c *decCoinValue) Type() string {
	return "cosmos.base.v1beta1.DecCoin"
}
