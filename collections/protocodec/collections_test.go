package protocodec_test

import (
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/collections/colltest"
	codec "cosmossdk.io/collections/protocodec"
)

func TestCollectionsCorrectness(t *testing.T) {
	t.Run("CollValueV2", func(t *testing.T) {
		// NOTE: we cannot use colltest.TestValueCodec because protov2 has different
		// compare semantics than protov1. We need to use protocmp.Transform() alongside
		// cmp to ensure equality.
		encoder := codec.CollValueV2[wrapperspb.UInt64Value]()
		value := &wrapperspb.UInt64Value{Value: 500}
		encodedValue, err := encoder.Encode(value)
		require.NoError(t, err)
		decodedValue, err := encoder.Decode(encodedValue)
		require.NoError(t, err)
		require.True(t, cmp.Equal(value, decodedValue, protocmp.Transform()), "encoding and decoding produces different values")

		encodedJSONValue, err := encoder.EncodeJSON(value)
		require.NoError(t, err)
		decodedJSONValue, err := encoder.DecodeJSON(encodedJSONValue)
		require.NoError(t, err)
		require.True(t, cmp.Equal(value, decodedJSONValue, protocmp.Transform()), "encoding and decoding produces different values")
		require.NotEmpty(t, encoder.ValueType())

		_ = encoder.Stringify(value)
	})

	t.Run("BoolValue", func(t *testing.T) {
		colltest.TestValueCodec(t, codec.BoolValue, true)
		colltest.TestValueCodec(t, codec.BoolValue, false)

		// asserts produced bytes are equal
		valueAssert := func(b bool) {
			wantBytes, err := (&gogotypes.BoolValue{Value: b}).Marshal()
			require.NoError(t, err)
			gotBytes, err := codec.BoolValue.Encode(b)
			require.NoError(t, err)
			require.Equal(t, wantBytes, gotBytes)
		}

		valueAssert(true)
		valueAssert(false)
	})
}
