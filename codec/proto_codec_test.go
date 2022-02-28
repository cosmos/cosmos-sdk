package codec_test

import (
	"errors"
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

func TestProtoMarshalInterface(t *testing.T) {
	cdc := codec.NewProtoCodec(createTestInterfaceRegistry())
	m := interfaceMarshaler{cdc.MarshalInterface, cdc.UnmarshalInterface}
	testInterfaceMarshaling(require.New(t), m, false)
	m = interfaceMarshaler{cdc.MarshalInterfaceJSON, cdc.UnmarshalInterfaceJSON}
	testInterfaceMarshaling(require.New(t), m, false)
}

func TestProtoCodec(t *testing.T) {
	cdc := codec.NewProtoCodec(createTestInterfaceRegistry())
	testMarshaling(t, cdc)
}

func TestProtoCodecUnmarshalLengthPrefixedChecks(t *testing.T) {
	cdc := codec.NewProtoCodec(createTestInterfaceRegistry())

	t.Run("Crafted bad uvarint size", func(t *testing.T) {
		crafted := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}
		recv := new(testdata.Cat)
		gotErr := cdc.UnmarshalLengthPrefixed(crafted, recv)
		require.Equal(t, gotErr, errors.New("invalid number of bytes read from length-prefixed encoding: -10"))

		require.Panics(t, func() { cdc.MustUnmarshalLengthPrefixed(crafted, recv) })
	})
}

func mustAny(msg proto.Message) *types.Any {
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return any
}

func BenchmarkProtoCodecMarshalLengthPrefixed(b *testing.B) {
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
		blob, err := pCdc.MarshalLengthPrefixed(msg)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(blob)))
	}
}

func TestGogoPulsarInterop(t *testing.T) {

}
