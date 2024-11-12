package flag

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type addressStringType struct{}

func (a addressStringType) NewValue(ctx *context.Context, b *Builder) Value {
	return &addressValue{addressCodec: b.AddressCodec, ctx: ctx}
}

func (a addressStringType) DefaultValue() string {
	return ""
}

type validatorAddressStringType struct{}

func (a validatorAddressStringType) NewValue(ctx *context.Context, b *Builder) Value {
	return &addressValue{addressCodec: b.ValidatorAddressCodec, ctx: ctx}
}

func (a validatorAddressStringType) DefaultValue() string {
	return ""
}

type addressValue struct {
	ctx          *context.Context
	addressCodec address.Codec

	cachedValue string
	value       string
}

func (a *addressValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if a.isEmpty() {
		return protoreflect.ValueOfString(""), nil
	}

	return a.get()
}

func (a *addressValue) String() string {
	return a.value
}

// Set implements the flag.Value interface for addressValue.
func (a *addressValue) Set(s string) error {
	_, err := a.addressCodec.StringToBytes(s)
	if err == nil {
		a.cachedValue = s
	}
	a.value = s

	return nil
}

func (a *addressValue) Type() string {
	return "account address or key name"
}

func (a *addressValue) isEmpty() bool {
	return a.value == "" && a.cachedValue == ""
}

func (a *addressValue) get() (protoreflect.Value, error) {
	if a.cachedValue != "" {
		return protoreflect.ValueOfString(a.cachedValue), nil
	}

	addr, err := getKey(a.value, a.addressCodec)
	if err != nil {
		return protoreflect.Value{}, err
	}
	a.cachedValue = addr

	return protoreflect.ValueOfString(addr), nil
}

type consensusAddressStringType struct{}

func (a consensusAddressStringType) NewValue(ctx *context.Context, b *Builder) Value {
	return &consensusAddressValue{
		addressValue: addressValue{
			addressCodec: b.ConsensusAddressCodec,
			ctx:          ctx,
		},
	}
}

func (a consensusAddressStringType) DefaultValue() string {
	return ""
}

type consensusAddressValue struct {
	addressValue
}

func (a consensusAddressValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	if a.isEmpty() {
		return protoreflect.ValueOfString(""), nil
	}

	addr, err := a.get()
	if err == nil {
		return addr, nil
	}

	// fallback to pubkey parsing
	registry := types.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	var pk cryptotypes.PubKey
	err2 := cdc.UnmarshalInterfaceJSON([]byte(a.value), &pk)
	if err2 != nil {
		return protoreflect.Value{}, fmt.Errorf("input isn't a pubkey (%w) or is an invalid account address (%w)", err, err2)
	}

	a.value, err = a.addressCodec.BytesToString(pk.Address())
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid pubkey address: %w", err)
	}
	a.cachedValue = a.value

	return protoreflect.ValueOfString(a.value), nil
}

func (a consensusAddressValue) String() string {
	return a.value
}

func (a *consensusAddressValue) Set(s string) error {
	_, err := a.addressCodec.StringToBytes(s)
	if err == nil {
		a.cachedValue = s
	}

	a.value = s

	return nil
}

func getKey(k string, ac address.Codec) (string, error) {
	if keybase != nil {
		addr, err := keybase.LookupAddressByKeyName(k)
		if err == nil {
			addrStr, err := ac.BytesToString(addr)
			if err != nil {
				return "", fmt.Errorf("invalid account address got from keyring: %w", err)
			}
			return addrStr, nil
		}
	}

	_, err := ac.StringToBytes(k)
	if err != nil {
		return "", fmt.Errorf("invalid account address or key name: %w", err)
	}

	return k, nil
}
