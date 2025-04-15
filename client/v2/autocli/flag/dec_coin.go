package flag

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/client/v2/internal/coins"
)

type decCoinType struct{}

type decCoinValue struct {
	value *basev1beta1.DecCoin
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
	return protoreflect.ValueOfMessage(c.value.ProtoReflect()), nil
}

func (c *decCoinValue) String() string {
	if c.value == nil {
		return ""
	}

	return c.value.String()
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
