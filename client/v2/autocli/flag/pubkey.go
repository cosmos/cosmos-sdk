package flag

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type pubkeyType struct{}

func (a pubkeyType) NewValue(_ context.Context, _ *Builder) Value {
	return pubkeyValue{}
}

func (a pubkeyType) DefaultValue() string {
	return ""
}

type pubkeyValue struct {
	value *types.Any
}

func (a pubkeyValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOf(a.value), nil
}

func (a pubkeyValue) String() string {
	return a.value.String()
}

func (a pubkeyValue) Set(s string) error {
	// fallback to pubkey parsing
	registry := types.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	var pk cryptotypes.PubKey
	err2 := cdc.UnmarshalInterfaceJSON([]byte(s), &pk)
	if err2 != nil {
		return fmt.Errorf("input isn't a pubkey, or is an invalid account address: %w", err2)
	}

	any, err := types.NewAnyWithValue(pk)
	if err != nil {
		return fmt.Errorf("error converting to any type")
	}
	a.value = any

	// a.value, err = a.addressCodec.BytesToString(pk.Address())
	// if err != nil {
	// 	return fmt.Errorf("invalid pubkey address: %w", err)
	// }

	return nil
}

func (a pubkeyValue) Type() string {
	return "pubkey"
}
