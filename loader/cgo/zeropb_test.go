package cgo

import (
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"pgregory.net/rapid"
)

func TestZeroPB(t *testing.T) {
	//gen := rapidproto.MessageGenerator(&bankv1beta1.MsgSend{}, rapidproto.GeneratorOptions{})
	rapid.Check(t, func(t *rapid.T) {
		//msg := gen.Draw(t, "msg")
		from := rapid.String().Draw(t, "from")
		to := rapid.String().Draw(t, "to")
		msg := &bankv1beta1.MsgSend{
			FromAddress: from,
			ToAddress:   to,
		}
		bz, err := ZeroPBMarshal(msg)
		require.NoError(t, err)

		msg2 := &bankv1beta1.MsgSend{}
		err = ZeroPBUnmarshal(bz, msg2)
		require.NoError(t, err)

		require.Empty(t, cmp.Diff(msg, msg2, protocmp.Transform()))
	})
}
