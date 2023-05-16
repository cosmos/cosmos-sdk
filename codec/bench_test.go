package codec_test

import (
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"github.com/stretchr/testify/require"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func BenchmarkLegacyGetSigners(b *testing.B) {
	_, _, addr := testdata.KeyTestPubAddr()
	msg := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   "",
		Amount:      nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.GetSigners()
	}
}

func BenchmarkProtoreflectGetSigners(b *testing.B) {
	cdc := codectestutil.CodecOptions{}.NewCodec()
	signingCtx := cdc.InterfaceRegistry().SigningContext()
	_, _, addr := testdata.KeyTestPubAddr()
	msg := &bankv1beta1.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   "",
		Amount:      nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := signingCtx.GetSigners(msg)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkProtoreflectGetSignersWithUnmarshal(b *testing.B) {
	cdc := codectestutil.CodecOptions{}.NewCodec()
	_, _, addr := testdata.KeyTestPubAddr()
	// start with a protoreflect message
	msg := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   "",
		Amount:      nil,
	}
	// marshal to an any first because this is what we get from the wire
	a, err := codectypes.NewAnyWithValue(msg)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := cdc.GetMsgAnySigners(a)
		if err != nil {
			panic(err)
		}
	}
}
