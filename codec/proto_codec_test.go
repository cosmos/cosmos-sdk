package codec_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

func TestProtoMarsharlInterface(t *testing.T) {
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

func TestEnsureRegistered(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	cat := &testdata.Cat{Moniker: "Garfield"}

	err := interfaceRegistry.EnsureRegistered(*cat)
	require.ErrorContains(t, err, "testdata.Cat is not a pointer")

	err = interfaceRegistry.EnsureRegistered(cat)
	require.ErrorContains(t, err, "testdata.Cat does not have a registered interface")

	interfaceRegistry.RegisterInterface("testdata.Animal",
		(*testdata.Animal)(nil),
		&testdata.Cat{},
	)

	require.NoError(t, interfaceRegistry.EnsureRegistered(cat))
}

func TestProtoCodecMarshal(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	interfaceRegistry.RegisterInterface("testdata.Animal",
		(*testdata.Animal)(nil),
		&testdata.Cat{},
	)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	cartonRegistry := types.NewInterfaceRegistry()
	cartonRegistry.RegisterInterface("testdata.Cartoon",
		(*testdata.Cartoon)(nil),
		&testdata.Bird{},
	)
	cartoonCdc := codec.NewProtoCodec(cartonRegistry)

	cat := &testdata.Cat{Moniker: "Garfield", Lives: 6}
	bird := &testdata.Bird{Species: "Passerina ciris"}
	require.NoError(t, interfaceRegistry.EnsureRegistered(cat))

	var (
		animal  testdata.Animal
		cartoon testdata.Cartoon
	)

	// sanity check
	require.True(t, reflect.TypeOf(cat).Implements(reflect.TypeOf((*testdata.Animal)(nil)).Elem()))

	bz, err := cdc.MarshalInterface(cat)
	require.NoError(t, err)

	err = cdc.UnmarshalInterface(bz, &animal)
	require.NoError(t, err)

	_, err = cdc.MarshalInterface(bird)
	require.ErrorContains(t, err, "does not have a registered interface")

	bz, err = cartoonCdc.MarshalInterface(bird)
	require.NoError(t, err)

	err = cdc.UnmarshalInterface(bz, &cartoon)
	require.ErrorContains(t, err, "no registered implementations")

	err = cartoonCdc.UnmarshalInterface(bz, &cartoon)
	require.NoError(t, err)

	// test typed nil input shouldn't panic
	var v *banktypes.QueryBalanceResponse
	bz, err = grpcServerEncode(cartoonCdc.GRPCCodec(), v)
	require.NoError(t, err)
	require.Empty(t, bz)
}

// Emulate grpc server implementation
// https://github.com/grpc/grpc-go/blob/b1d7f56b81b7902d871111b82dec6ba45f854ede/rpc_util.go#L590
func grpcServerEncode(c encoding.Codec, msg interface{}) ([]byte, error) {
	if msg == nil { // NOTE: typed nils will not be caught by this check
		return nil, nil
	}
	b, err := c.Marshal(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "grpc: error while marshaling: %v", err.Error())
	}
	if uint(len(b)) > math.MaxUint32 {
		return nil, status.Errorf(codes.ResourceExhausted, "grpc: message too large (%d bytes)", len(b))
	}
	return b, nil
}

func mustAny(msg proto.Message) *types.Any {
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return any
}

func BenchmarkProtoCodecMarshalLengthPrefixed(b *testing.B) {
	pCdc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	msg := &testdata.HasAnimal{
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

func TestGetSigners(t *testing.T) {
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		SigningOptions: signing.Options{
			AddressCodec:          testAddressCodec{},
			ValidatorAddressCodec: testAddressCodec{},
		},
		ProtoFiles: protoregistry.GlobalFiles,
	})
	require.NoError(t, err)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	testAddr := sdk.AccAddress("test")
	testAddrStr := testAddr.String()
	testAddr2 := sdk.AccAddress("test2")
	testAddrStr2 := testAddr2.String()

	msgSendV1 := banktypes.NewMsgSend(testAddr, testAddr2, sdk.NewCoins(sdk.NewCoin("foo", sdkmath.NewInt(1))))
	msgSendV2 := &bankv1beta1.MsgSend{
		FromAddress: testAddrStr,
		ToAddress:   testAddrStr2,
		Amount:      []*basev1beta1.Coin{{Denom: "foo", Amount: "1"}},
	}

	signers, msgSendV2Copy, err := cdc.GetMsgV1Signers(msgSendV1)
	require.NoError(t, err)
	require.Equal(t, [][]byte{testAddr}, signers)
	require.True(t, protov2.Equal(msgSendV2, msgSendV2Copy))

	signers, err = cdc.GetMsgV2Signers(msgSendV2)
	require.NoError(t, err)
	require.Equal(t, [][]byte{testAddr}, signers)

	msgSendAny, err := types.NewAnyWithValue(msgSendV1)
	require.NoError(t, err)
	signers, msgSendV2Copy, err = cdc.GetMsgAnySigners(msgSendAny)
	require.NoError(t, err)
	require.Equal(t, [][]byte{testAddr}, signers)
	require.True(t, protov2.Equal(msgSendV2, msgSendV2Copy))
}

type testAddressCodec struct{}

func (t testAddressCodec) StringToBytes(text string) ([]byte, error) {
	return sdk.AccAddressFromBech32(text)
}

func (t testAddressCodec) BytesToString(bz []byte) (string, error) {
	return sdk.AccAddress(bz).String(), nil
}
