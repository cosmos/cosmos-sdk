package signing

import (
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestGetSigners(t *testing.T) {
	ctx := MsgContextOptions{}.Build()
	signers, err := ctx.GetSignersForMessage(&bankv1beta1.MsgSend{
		FromAddress: "foo",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"foo"}, signers)

	signers, err = ctx.GetSignersForMessage(&bankv1beta1.MsgMultiSend{
		Inputs: []*bankv1beta1.Input{
			{Address: "foo"},
			{Address: "bar"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar"}, signers)
}
