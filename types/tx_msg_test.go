package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/stretchr/testify/require"

	sdk "github.com/KiraCore/cosmos-sdk/types"
)

func TestTestMsg(t *testing.T) {
	t.Parallel()
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	accAddr := sdk.AccAddress(addr)

	msg := testdata.NewTestMsg(accAddr)
	require.NotNil(t, msg)
	require.Equal(t, "TestMsg", msg.Route())
	require.Equal(t, "Test message", msg.Type())
	require.Nil(t, msg.ValidateBasic())
	require.NotPanics(t, func() { msg.GetSignBytes() })
	require.Equal(t, []sdk.AccAddress{accAddr}, msg.GetSigners())
}
