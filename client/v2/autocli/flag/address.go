package flag

import (
	"context"
	"fmt"

	reflectionv2alpha1 "cosmossdk.io/api/cosmos/base/reflection/v2alpha1"
	"cosmossdk.io/core/address"
	"google.golang.org/protobuf/reflect/protoreflect"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
)

type addressStringType struct{}

func (a addressStringType) NewValue(ctx context.Context, b *Builder) Value {
	if b.AddressCodec == nil {
		conn, err := b.GetClientConn()
		if err != nil {
			panic(err)
		}
		reflectionClient := reflectionv2alpha1.NewReflectionServiceClient(conn)
		resp, err := reflectionClient.GetConfigurationDescriptor(ctx, &reflectionv2alpha1.GetConfigurationDescriptorRequest{})
		if err != nil {
			panic(err)
		}
		if resp == nil || resp.Config == nil {
			panic("bech32 account address prefix is not set")
		}

		b.AddressCodec = addresscodec.NewBech32Codec(resp.Config.Bech32AccountAddressPrefix)
	}

	return &addressValue{addressCodec: b.AddressCodec}
}

func (a addressStringType) DefaultValue() string {
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
