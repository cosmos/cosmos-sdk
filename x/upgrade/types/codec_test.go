package types

import (
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/gov/types/v1beta1"
	"cosmossdk.io/x/tx/signing"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

func TestInterfaceRegistrationOfContent(t *testing.T) {
	opts := codectestutil.CodecOptions{}
	registrar, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          opts.GetAddressCodec(),
			ValidatorAddressCodec: opts.GetValidatorCodec(),
		},
	})
	require.NoError(t, err)
	RegisterInterfaces(registrar)
	val := &gogoprotoany.Any{
		TypeUrl: "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
		Value:   []byte{},
	}
	require.NoError(t, registrar.UnpackAny(val, new(v1beta1.Content)))
}
