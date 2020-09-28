package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgTestSuite(t *testing.T) {
	t.Parallel()
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	accAddr := sdk.AccAddress(addr)

	msg := testdata.NewTestMsg(accAddr)
	require.NotNil(t, msg)
	require.Equal(t, []sdk.AccAddress{accAddr}, msg.GetSigners())
	require.Equal(t, "TestMsg", msg.Route())
	require.Equal(t, "Test message", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.NotPanics(t, func() { msg.GetSignBytes() })
}
