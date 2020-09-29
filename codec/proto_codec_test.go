package codec_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func createTestInterfaceRegistry() types.InterfaceRegistry {
	interfaceRegistry := types.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("testdata.Animal",
		(*testdata.Animal)(nil),
		&testdata.Dog{},
		&testdata.Cat{},
	)

	return interfaceRegistry
}

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
			codec.NewProtoCodec(createTestInterfaceRegistry()),
			&testdata.Dog{Name: "rufus"},
			&testdata.Dog{},
			false,
			false,
		},
		{
			"invalid decode type",
			codec.NewProtoCodec(createTestInterfaceRegistry()),
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

type lyingProtoMarshaler struct {
	codec.ProtoMarshaler
	falseSize int
}

func (lpm *lyingProtoMarshaler) Size() int {
	return lpm.falseSize
}

func TestProtoCodecUnmarshalBinaryLengthPrefixedChecks(t *testing.T) {
	cdc := codec.NewProtoCodec(createTestInterfaceRegistry())

	truth := &testdata.Cat{Lives: 9, Moniker: "glowing"}
	realSize := len(cdc.MustMarshalBinaryBare(truth))

	falseSizes := []int{
		100,
		5,
	}

	for _, falseSize := range falseSizes {
		falseSize := falseSize

		t.Run(fmt.Sprintf("ByMarshaling falseSize=%d", falseSize), func(t *testing.T) {
			lpm := &lyingProtoMarshaler{
				ProtoMarshaler: &testdata.Cat{Lives: 9, Moniker: "glowing"},
				falseSize:      falseSize,
			}
			var serialized []byte
			require.NotPanics(t, func() { serialized = cdc.MustMarshalBinaryLengthPrefixed(lpm) })

			recv := new(testdata.Cat)
			gotErr := cdc.UnmarshalBinaryLengthPrefixed(serialized, recv)
			var wantErr error
			if falseSize > realSize {
				wantErr = fmt.Errorf("not enough bytes to read; want: %d, got: %d", falseSize, realSize)
			} else {
				wantErr = fmt.Errorf("too many bytes to read; want: %d, got: %d", falseSize, realSize)
			}
			require.Equal(t, gotErr, wantErr)
		})
	}

	t.Run("Crafted bad uvarint size", func(t *testing.T) {
		crafted := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}
		recv := new(testdata.Cat)
		gotErr := cdc.UnmarshalBinaryLengthPrefixed(crafted, recv)
		require.Equal(t, gotErr, errors.New("invalid number of bytes read from length-prefixed encoding: -10"))

		require.Panics(t, func() { cdc.MustUnmarshalBinaryLengthPrefixed(crafted, recv) })
	})
}

func mustAny(msg proto.Message) *types.Any {
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return any
}

func BenchmarkProtoCodecMarshalBinaryLengthPrefixed(b *testing.B) {
	var pCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
	var msg = &testdata.HasAnimal{
		X: 1000,
		Animal: mustAny(&testdata.HasAnimal{
			X: 2000,
			Animal: mustAny(&testdata.HasAnimal{
				X: 3000,
				Animal: mustAny(&testdata.HasAnimal{
					X: 4000,
					Animal: mustAny(&testdata.HasAnimal{
						X: 5000,
						Animal: mustAny(&testdata.Cat{
							Moniker: "Garfield",
							Lives:   6,
						}),
					}),
				}),
			}),
		}),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		blob, err := pCdc.MarshalBinaryLengthPrefixed(msg)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(blob)))
	}
}
