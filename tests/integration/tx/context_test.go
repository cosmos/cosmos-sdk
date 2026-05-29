package tx

import (
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/pulsar/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
)

func TestDefineCustomGetSigners(t *testing.T) {
	msg := &testpb.TestRepeatedFields{
		NullableDontOmitempty: []*testpb.Streng{
			{Value: "foo"},
			{Value: "bar"},
		},
	}

	// With a custom GetSigners for TestRepeatedFields, the signer should be "bar".
	signingOpts := txsigning.Options{
		AddressCodec:          address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()},
		ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix()},
	}
	signingOpts.DefineCustomGetSigners(
		proto.MessageName(&testpb.TestRepeatedFields{}),
		func(msg proto.Message) ([][]byte, error) {
			testMsg := msg.(*testpb.TestRepeatedFields)
			signer := testMsg.NullableDontOmitempty[1].Value
			return [][]byte{[]byte(signer)}, nil
		},
	)

	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles:     gogoproto.HybridResolver,
		SigningOptions: signingOpts,
	})
	require.NoError(t, err)
	require.NotNil(t, ir)

	signers, err := ir.SigningContext().GetSigners(msg)
	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("bar")}, signers)

	// Without the custom signer registered, GetSigners should fail.
	irDefault, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix()},
			ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix()},
		},
	})
	require.NoError(t, err)

	_, err = irDefault.SigningContext().GetSigners(msg)
	require.ErrorContains(t, err, "use DefineCustomGetSigners")
}
