package baseapp

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	allowedMsgs := []sdk.Msg{&testdata.TestMsg{Signers: []string{"sender"}}}
	cb := NewCircuitBreaker(allowedMsgs)

	ctx := sdk.Context{}
	msg := &testdata.TestMsg{
		Signers: []string{"sender"},
	}
	require.True(t, cb.IsAllowed(ctx, msg))

	cb.AllowedMsgs[sdk.MsgTypeURL(msg)] = false
	require.False(t, cb.IsAllowed(ctx, msg))
}
