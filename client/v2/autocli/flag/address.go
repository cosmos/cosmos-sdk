package flag

import (
	"context"

	reflectionv2alpha1 "cosmossdk.io/api/cosmos/base/reflection/v2alpha1"
	"github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type addressStringType struct{}

func (a addressStringType) NewValue(ctx context.Context, b *Builder) Value {
	if b.AddressPrefix == "" {
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
		b.AddressPrefix = resp.Config.Bech32AccountAddressPrefix
	}
	return &addressValue{addressPrefix: b.AddressPrefix}
}

func (a addressStringType) DefaultValue() string {
	return ""
}

type addressValue struct {
	value         string
	addressPrefix string
}

func (a addressValue) Get(protoreflect.Value) (protoreflect.Value, error) {
	return protoreflect.ValueOfString(a.value), nil
}

func (a addressValue) String() string {
	return a.value
}

// Set implements the flag.Value interface for addressValue it only supports bech32 addresses.
func (a *addressValue) Set(s string) error {
	_, err := types.GetFromBech32(s, a.addressPrefix)
	if err != nil {
		return err
	}
	a.value = s

	return nil
}

func (a addressValue) Type() string {
	return "bech32 account address key name"
}
