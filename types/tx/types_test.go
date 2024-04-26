package tx

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	types2 "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTx_GetMsgs(t *testing.T) {
	any1, err := types2.NewAnyWithCacheWithValue(&DummyProtoMessage1{})
	require.Nil(t, err)
	any2, err := types2.NewAnyWithCacheWithValue(&DummyProtoMessage2{})
	require.Nil(t, err)

	cases := []struct {
		name     string
		msgs     []*types2.Any
		expected []sdk.Msg
		expPanic bool
	}{
		{
			name:     "Tx with one message",
			msgs:     []*types2.Any{any1},
			expected: []sdk.Msg{&DummyProtoMessage1{}},
			expPanic: false,
		},
		{
			name:     "Tx with messages of the same type",
			msgs:     []*types2.Any{any1, any1},
			expected: []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage1{}},
			expPanic: false,
		},
		{
			name:     "Tx with messages with different type",
			msgs:     []*types2.Any{any1, any2},
			expected: []sdk.Msg{&DummyProtoMessage1{}, &DummyProtoMessage2{}},
			expPanic: false,
		},
		{
			name:     "Tx without uncached messages",
			msgs:     []*types2.Any{{}},
			expPanic: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transaction := Tx{
				Body: &TxBody{
					Messages: tc.msgs,
				},
			}

			if tc.expPanic {
				require.Panics(t, func() {
					transaction.GetMsgs()
				})
				return
			}
			actual := transaction.GetMsgs()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestTx_ValidateBasic(t *testing.T) {
	cases := []struct {
		name        string
		transaction *Tx
		expErr      bool
	}{
		{
			name:        "Tx is nil",
			transaction: nil,
			expErr:      true,
		},
		{
			name:        "Tx without body",
			transaction: &Tx{},
			expErr:      true,
		},
		{
			name:        "Tx without AuthInfo",
			transaction: &Tx{Body: &TxBody{}},
			expErr:      true,
		},
		{
			name:        "Tx without Fee",
			transaction: &Tx{Body: &TxBody{}, AuthInfo: &AuthInfo{}},
			expErr:      true,
		},
		{
			name:        "Tx with gas limit greater than Max gas wanted",
			transaction: &Tx{Body: &TxBody{}, AuthInfo: &AuthInfo{Fee: &Fee{GasLimit: MaxGasWanted + 1}}},
			expErr:      true,
		},
		{
			name:        "Tx without Fee Amount",
			transaction: &Tx{Body: &TxBody{}, AuthInfo: &AuthInfo{Fee: &Fee{GasLimit: MaxGasWanted}}},
			expErr:      true,
		},
		{
			name: "Tx with negative Fee Amount",
			transaction: &Tx{
				Body: &TxBody{},
				AuthInfo: &AuthInfo{
					Fee: &Fee{GasLimit: MaxGasWanted, Amount: sdk.Coins{sdk.Coin{Amount: math.NewInt(-1)}}},
				},
			},
			expErr: true,
		},
		{
			name: "Tx with invalid fee payer address",
			transaction: &Tx{
				Body: &TxBody{},
				AuthInfo: &AuthInfo{
					Fee: &Fee{
						GasLimit: MaxGasWanted,
						Payer:    "invalidPayerAddress",
						Amount:   sdk.Coins{sdk.NewCoin("aaa", math.NewInt(10))},
					},
				},
			},
			expErr: true,
		},
		{
			name: "Tx without signature",
			transaction: &Tx{
				Body: &TxBody{},
				AuthInfo: &AuthInfo{
					Fee: &Fee{
						GasLimit: MaxGasWanted,
						Payer:    "cosmos1ulav3hsenupswqfkw2y3sup5kgtqwnvqa8eyhs",
						Amount:   sdk.Coins{sdk.NewCoin("aaa", math.NewInt(11))},
					},
				},
			},
			expErr: true,
		},
		{
			name: "Tx is valid",
			transaction: &Tx{
				Body: &TxBody{},
				AuthInfo: &AuthInfo{
					Fee: &Fee{
						GasLimit: MaxGasWanted,
						Payer:    "cosmos1ulav3hsenupswqfkw2y3sup5kgtqwnvqa8eyhs",
						Amount:   sdk.Coins{sdk.NewCoin("aaa", math.NewInt(11))},
					},
				},
				Signatures: [][]byte{[]byte("signature")},
			},
			expErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.transaction.ValidateBasic()
			require.Equal(t, tc.expErr, err != nil)
		})
	}
}

func TestTx_GetSigners(t *testing.T) {
	transaction := &Tx{
		Body: &TxBody{},
		AuthInfo: &AuthInfo{
			Fee: &Fee{
				GasLimit: MaxGasWanted,
				Payer:    "cosmos1ulav3hsenupswqfkw2y3sup5kgtqwnvqa8eyhs",
				Amount:   sdk.Coins{sdk.NewCoin("aaa", math.NewInt(11))},
			},
		},
		Signatures: [][]byte{[]byte("signature")},
	}

	addrCdc := address.NewBech32Codec("cosmos")
	options := codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          addrCdc,
			ValidatorAddressCodec: dummyAddressCodec{},
		},
	}
	ir, err := codectypes.NewInterfaceRegistryWithOptions(options)
	require.Nil(t, err)
	cdc := codec.NewProtoCodec(ir)
	b, _, err := transaction.GetSigners(cdc)
	require.Nil(t, err)

	expect := "cosmos1ulav3hsenupswqfkw2y3sup5kgtqwnvqa8eyhs"
	actual, err := addrCdc.BytesToString(b[0])
	require.Equal(t, expect, actual)
	require.Nil(t, err)
}

type DummyProtoMessage1 struct{}

func (d *DummyProtoMessage1) Reset()         {}
func (d *DummyProtoMessage1) String() string { return "/dummy.proto.message1" }
func (d *DummyProtoMessage1) ProtoMessage()  {}

type DummyProtoMessage2 struct{}

func (d *DummyProtoMessage2) Reset()         {}
func (d *DummyProtoMessage2) String() string { return "/dummy.proto.message2" }
func (d *DummyProtoMessage2) ProtoMessage()  {}

type dummyAddressCodec struct{}

func (d dummyAddressCodec) StringToBytes(text string) ([]byte, error) { return hex.DecodeString(text) }
func (d dummyAddressCodec) BytesToString(bz []byte) (string, error) {
	return hex.EncodeToString(bz), nil
}
