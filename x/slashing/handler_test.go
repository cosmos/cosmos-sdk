package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestCannotUnjailUnlessJailed(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t)
	slh := NewHandler(keeper)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	msg := newTestMsgCreateValidator(addr, val, amt)
	got := stake.NewHandler(sk)(ctx, msg)
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))

	// assert non-jailed validator can't be unjailed
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK(), "allowed unjail of non-jailed validator")
	require.Equal(t, sdk.ToABCICode(DefaultCodespace, CodeValidatorNotJailed), got.Code)
}
