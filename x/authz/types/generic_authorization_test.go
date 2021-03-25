package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func TestGenericAuthorization(t *testing.T) {

	// validateBasic returns error not a valida service msg
	authorization := types.NewGenericAuthorization(banktypes.TypeMsgSend)
	require.Error(t, authorization.ValidateBasic())

	// valid service msg
	authorization = types.NewGenericAuthorization(banktypes.SendAuthorization{}.MethodName())
	require.NoError(t, authorization.ValidateBasic())
	require.Equal(t, banktypes.SendAuthorization{}.MethodName(), authorization.MessageName)
}
