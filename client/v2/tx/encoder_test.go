package tx

import (
	"testing"

	"github.com/stretchr/testify/require"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
)

func getWrappedTx(t *testing.T) *wrappedTx {
	t.Helper()

	pk := secp256k1.GenPrivKey().PubKey()
	addr, _ := ac.BytesToString(pk.Address())
	b := newTxBuilder(ac, decoder, cdc)

	err := b.SetMsgs([]transaction.Msg{&countertypes.MsgIncreaseCounter{
		Signer: addr,
		Count:  0,
	}}...)
	require.NoError(t, err)

	err = b.SetFeePayer(addr)
	require.NoError(t, err)

	b.SetFeeAmount([]*base.Coin{{
		Denom:  "cosmos",
		Amount: "1000",
	}})

	err = b.SetSignatures([]Signature{{
		PubKey: pk,
		Data: &SingleSignatureData{
			SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: 0,
	}}...)
	require.NoError(t, err)
	wTx, err := b.getTx()
	return wTx
}

func Test_txEncoder_txDecoder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "encode tx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wTx := getWrappedTx(t)

			encodedTx, err := encodeTx(wTx)
			require.NoError(t, err)
			require.NotNil(t, encodedTx)

			isDeterministic, err := encodeTx(wTx)
			require.NoError(t, err)
			require.NotNil(t, encodedTx)
			require.Equal(t, encodedTx, isDeterministic)

			//decodedTx, err := decodeTx(encodedTx)
			//require.NoError(t, err)
			//require.NotNil(t, decodedTx)
		})
	}
}

func Test_txJsonEncoder_txJsonDecoder(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "json encode and decode tx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wTx := getWrappedTx(t)

			encodedTx, err := encodeJsonTx(wTx)
			require.NoError(t, err)
			require.NotNil(t, encodedTx)

			decodedTx, err := decodeJsonTx(encodedTx)
			require.NoError(t, err)
			require.NotNil(t, decodedTx)
		})
	}
}
