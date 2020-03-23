package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testdata"
)

func TestProtoCodec(t *testing.T) {
	testCases := []struct {
		name         string
		codec        codec.Marshaler
		input        codec.ProtoMarshaler
		recv         codec.ProtoMarshaler
		marshalErr   bool
		unmarshalErr bool
	}{
		{
			"valid encoding and decoding",
			codec.NewProtoCodec(),
			&testdata.Dog{Name: "rufus"},
			&testdata.Dog{},
			false,
			false,
		},
		{
			"invalid decode type",
			codec.NewProtoCodec(),
			&testdata.Dog{Name: "rufus"},
			&testdata.Cat{},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			bz, err := tc.codec.MarshalBinaryBare(tc.input)

			if tc.marshalErr {
				require.Error(t, err)
				require.Panics(t, func() { tc.codec.MustMarshalBinaryBare(tc.input) })
			} else {
				var bz2 []byte
				require.NoError(t, err)
				require.NotPanics(t, func() { bz2 = tc.codec.MustMarshalBinaryBare(tc.input) })
				require.Equal(t, bz, bz2)

				err := tc.codec.UnmarshalBinaryBare(bz, tc.recv)
				if tc.unmarshalErr {
					require.Error(t, err)
					require.Panics(t, func() { tc.codec.MustUnmarshalBinaryBare(bz, tc.recv) })
				} else {
					require.NoError(t, err)
					require.NotPanics(t, func() { tc.codec.MustUnmarshalBinaryBare(bz, tc.recv) })
					require.Equal(t, tc.input, tc.recv)
				}
			}

			bz, err = tc.codec.MarshalBinaryLengthPrefixed(tc.input)
			if tc.marshalErr {
				require.Error(t, err)
				require.Panics(t, func() { tc.codec.MustMarshalBinaryLengthPrefixed(tc.input) })
			} else {
				var bz2 []byte
				require.NoError(t, err)
				require.NotPanics(t, func() { bz2 = tc.codec.MustMarshalBinaryLengthPrefixed(tc.input) })
				require.Equal(t, bz, bz2)

				err := tc.codec.UnmarshalBinaryLengthPrefixed(bz, tc.recv)
				if tc.unmarshalErr {
					require.Error(t, err)
					require.Panics(t, func() { tc.codec.MustUnmarshalBinaryLengthPrefixed(bz, tc.recv) })
				} else {
					require.NoError(t, err)
					require.NotPanics(t, func() { tc.codec.MustUnmarshalBinaryLengthPrefixed(bz, tc.recv) })
					require.Equal(t, tc.input, tc.recv)
				}
			}

			bz, err = tc.codec.MarshalJSON(tc.input)
			if tc.marshalErr {
				require.Error(t, err)
				require.Panics(t, func() { tc.codec.MustMarshalJSON(tc.input) })
			} else {
				var bz2 []byte
				require.NoError(t, err)
				require.NotPanics(t, func() { bz2 = tc.codec.MustMarshalJSON(tc.input) })
				require.Equal(t, bz, bz2)

				err := tc.codec.UnmarshalJSON(bz, tc.recv)
				if tc.unmarshalErr {
					require.Error(t, err)
					require.Panics(t, func() { tc.codec.MustUnmarshalJSON(bz, tc.recv) })
				} else {
					require.NoError(t, err)
					require.NotPanics(t, func() { tc.codec.MustUnmarshalJSON(bz, tc.recv) })
					require.Equal(t, tc.input, tc.recv)
				}
			}
		})
	}
}
