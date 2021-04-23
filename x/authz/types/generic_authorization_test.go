package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestGenericAuthorization(t *testing.T) {
	t.Log("verify ValidateBasic returns error for non-service msg")
	authorization := types.NewGenericAuthorization(banktypes.TypeMsgSend)
	require.Error(t, authorization.ValidateBasic())

	t.Log("verify ValidateBasic returns nil for service msg")
	authorization = types.NewGenericAuthorization(banktypes.SendAuthorization{}.MsgTypeURL())
	require.NoError(t, authorization.ValidateBasic())
	require.Equal(t, banktypes.SendAuthorization{}.MsgTypeURL(), authorization.Msg)
}
