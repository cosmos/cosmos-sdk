package codec_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func createTestCodec() *codec.LegacyAmino {
	cdc := codec.New()

	cdc.RegisterInterface((*testdata.Animal)(nil), nil)
	cdc.RegisterConcrete(testdata.Dog{}, "testdata/Dog", nil)
	cdc.RegisterConcrete(testdata.Cat{}, "testdata/Cat", nil)

	return cdc
}

func TestAminoCodec(t *testing.T) {
	any, err := types.NewAnyWithValue(&testdata.Dog{Name: "rufus"})
	require.NoError(t, err)

	testCases := []struct {
		name         string
		codec        *codec.AminoCodec
		input        codec.ProtoMarshaler
		recv         codec.ProtoMarshaler
		marshalErr   bool
		unmarshalErr bool
	}{
		{
			"valid encoding and decoding",
			codec.NewAminoCodec(createTestCodec()),
			&testdata.Dog{Name: "rufus"},
			&testdata.Dog{},
			false,
			false,
		},
		{
			"invalid decode type",
			codec.NewAminoCodec(createTestCodec()),
			&testdata.Dog{Name: "rufus"},
			&testdata.Cat{},
			false,
			true,
		},
		{
			"any marshaling",
			codec.NewAminoCodec(createTestCodec()),
			&testdata.HasAnimal{Animal: any},
			&testdata.HasAnimal{Animal: any},
			false,
			false,
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

func TestAminoCodecMarshalJSONIndent(t *testing.T) {
	any, err := types.NewAnyWithValue(&testdata.Dog{Name: "rufus"})
	require.NoError(t, err)

	testCases := []struct {
		name       string
		input      interface{}
		marshalErr bool
		wantJSON   string
	}{
		{
			name:  "valid encoding and decoding",
			input: &testdata.Dog{Name: "rufus"},
			wantJSON: `{
  "type": "testdata/Dog",
  "value": {
    "name": "rufus"
  }
}`,
		},
		{
			name:  "any marshaling",
			input: &testdata.HasAnimal{Animal: any},
			wantJSON: `{
  "animal": {
    "type": "testdata/Dog",
    "value": {
      "name": "rufus"
    }
  }
}`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cdc := codec.NewAminoCodec(createTestCodec())
			bz, err := cdc.MarshalJSONIndent(tc.input, "", "  ")

			if tc.marshalErr {
				require.Error(t, err)
				require.Panics(t, func() { codec.MustMarshalJSONIndent(cdc.LegacyAmino, tc.input) })
				return
			}

			// Otherwise these are expected to pass.
			require.NoError(t, err)
			require.Equal(t, bz, []byte(tc.wantJSON))

			var bz2 []byte
			require.NotPanics(t, func() { bz2 = codec.MustMarshalJSONIndent(cdc.LegacyAmino, tc.input) })
			require.Equal(t, bz2, []byte(tc.wantJSON))
		})
	}
}

func TestAminoCodecPrintTypes(t *testing.T) {
	cdc := codec.NewAminoCodec(createTestCodec())
	buf := new(bytes.Buffer)
	require.NoError(t, cdc.PrintTypes(buf))
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	require.True(t, len(lines) > 1)
	wantHeader := "| Type | Name | Prefix | Length | Notes |"
	require.Equal(t, lines[0], []byte(wantHeader))

	// Expecting the types to be listed in the order that they were registered.
	require.True(t, bytes.HasPrefix(lines[2], []byte("| Dog | testdata/Dog |")))
	require.True(t, bytes.HasPrefix(lines[3], []byte("| Cat | testdata/Cat |")))
}

func TestAminoCodecUnpackAnyFails(t *testing.T) {
	cdc := codec.NewAminoCodec(createTestCodec())
	err := cdc.UnpackAny(new(types.Any), &testdata.Cat{})
	require.Error(t, err)
	require.Equal(t, err, errors.New("AminoCodec can't handle unpack protobuf Any's"))
}
