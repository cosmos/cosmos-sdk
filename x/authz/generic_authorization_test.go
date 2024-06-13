package authz_test

import (
	"testing"

	"cosmossdk.io/x/authz"
	banktypes "cosmossdk.io/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGenericAuthorization(t *testing.T) {
	t.Log("verify ValidateBasic returns nil for service msg")
	a := authz.NewGenericAuthorization(banktypes.SendAuthorization{}.MsgTypeURL())
	require.NoError(t, a.ValidateBasic())
	require.Equal(t, banktypes.SendAuthorization{}.MsgTypeURL(), a.Msg)
}

func TestGenericSpendAuthorization(t *testing.T) {
	t.Log("verify ValidateBasic returns nil for service msg")

	coin, _ := sdk.ParseCoinNormalized("1000ustake")
	a := authz.NewGenericAuthorizationWithSpendLimit(banktypes.SendAuthorization{}.MsgTypeURL(), sdk.NewCoins(coin))
	require.NoError(t, a.ValidateBasic())
	require.Equal(t, banktypes.SendAuthorization{}.MsgTypeURL(), a.Msg)
}
