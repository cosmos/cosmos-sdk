package decode_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-proto/anyutil"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/decode"
	"cosmossdk.io/x/tx/internal/testpb"
	"cosmossdk.io/x/tx/signing"
)

type mockCodec struct{}

func (m mockCodec) Unmarshal(bytes []byte, message gogoproto.Message) error {
	return gogoproto.Unmarshal(bytes, message)
}

func TestDecode(t *testing.T) {
	accSeq := uint64(2)

	pkAny, err := anyutil.New(&secp256k1.PubKey{Key: []byte("foo")})
	require.NoError(t, err)
	var signerInfo []*txv1beta1.SignerInfo
	signerInfo = append(signerInfo, &txv1beta1.SignerInfo{
		PublicKey: pkAny,
		ModeInfo: &txv1beta1.ModeInfo{
			Sum: &txv1beta1.ModeInfo_Single_{
				Single: &txv1beta1.ModeInfo_Single{
					Mode: signingv1beta1.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
		Sequence: accSeq,
	})

	signingCtx, err := signing.NewContext(signing.Options{
		AddressCodec:          dummyAddressCodec{},
		ValidatorAddressCodec: dummyAddressCodec{},
	})
	require.NoError(t, err)
	decoder, err := decode.NewDecoder(decode.Options{
		SigningContext: signingCtx,
		ProtoCodec:     mockCodec{},
	})
	require.NoError(t, err)

	gogoproto.RegisterType(&bankv1beta1.MsgSend{}, string((&bankv1beta1.MsgSend{}).ProtoReflect().Descriptor().FullName()))
	gogoproto.RegisterType(&testpb.A{}, string((&testpb.A{}).ProtoReflect().Descriptor().FullName()))

	testCases := []struct {
		name            string
		msg             proto.Message
		feePayer        string
		error           string
		expectedSigners int
	}{
		{
			name:            "happy path",
			msg:             &bankv1beta1.MsgSend{},
			expectedSigners: 1,
		},
		{
			name:  "empty signer option",
			msg:   &testpb.A{},
			error: "no cosmos.msg.v1.signer option found for message A; use DefineCustomGetSigners to specify a custom getter: tx parse error",
		},
		{
			name:     "invalid feePayer",
			msg:      &bankv1beta1.MsgSend{},
			feePayer: "payer",
			error:    `encoding/hex: invalid byte: U+0070 'p': tx parse error`,
		},
		{
			name:            "valid feePayer",
			msg:             &bankv1beta1.MsgSend{},
			feePayer:        "636f736d6f733168363935356b3836397a72306770383975717034337a373263393033666d35647a366b75306c", // hexadecimal to work with dummyAddressCodec
			expectedSigners: 2,
		},
		{
			name: "same msg signer and feePayer",
			msg: &bankv1beta1.MsgSend{
				FromAddress: "636f736d6f733168363935356b3836397a72306770383975717034337a373263393033666d35647a366b75306c",
			},
			feePayer:        "636f736d6f733168363935356b3836397a72306770383975717034337a373263393033666d35647a366b75306c",
			expectedSigners: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := proto.Marshal(tc.msg)
			require.NoError(t, err)

			anyMsg, err := anyutil.New(tc.msg)
			require.NoError(t, err)
			tx := &txv1beta1.Tx{
				Body: &txv1beta1.TxBody{
					Messages:      []*anypb.Any{anyMsg},
					Memo:          "memo",
					TimeoutHeight: 0,
				},
				AuthInfo: &txv1beta1.AuthInfo{
					SignerInfos: signerInfo,
					Fee: &txv1beta1.Fee{
						Amount:   []*basev1beta1.Coin{{Amount: "100", Denom: "denom"}},
						GasLimit: 100,
						Payer:    tc.feePayer,
						Granter:  "",
					},
				},
				Signatures: nil,
			}
			txBytes, err := proto.Marshal(tx)
			require.NoError(t, err)

			decodeTx, err := decoder.Decode(txBytes)
			if tc.error != "" {
				require.EqualError(t, err, tc.error)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(decodeTx.Signers), tc.expectedSigners)

			require.Equal(t,
				fmt.Sprintf("/%s", tc.msg.ProtoReflect().Descriptor().FullName()),
				decodeTx.Tx.Body.Messages[0].TypeUrl)
		})
	}
}

type dummyAddressCodec struct{}

func (d dummyAddressCodec) StringToBytes(text string) ([]byte, error) {
	return hex.DecodeString(text)
}

func (d dummyAddressCodec) BytesToString(bz []byte) (string, error) {
	return hex.EncodeToString(bz), nil
}

func TestDecodeTxBodyPanic(t *testing.T) {
	crashVector := []byte{
		0x0a, 0x0a, 0x09, 0xe7, 0xbf, 0xba, 0xe6, 0x82, 0x9a, 0xe6, 0xaa, 0x30,
	}

	cdc := new(dummyAddressCodec)
	signingCtx, err := signing.NewContext(signing.Options{
		AddressCodec:          cdc,
		ValidatorAddressCodec: cdc,
	})
	if err != nil {
		t.Fatal(err)
	}
	dec, err := decode.NewDecoder(decode.Options{
		SigningContext: signingCtx,
		ProtoCodec:     mockCodec{},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = dec.Decode(crashVector)
	if err == nil {
		t.Fatal("expected a non-nil error")
	}
	if g, w := err.Error(), "could not consume length prefix"; !strings.Contains(g, w) {
		t.Fatalf("error mismatch\n%s\nodes not contain\n\t%q", g, w)
	}
}
