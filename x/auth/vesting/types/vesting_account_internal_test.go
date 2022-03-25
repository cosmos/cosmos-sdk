package types

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	tmtime "github.com/tendermint/tendermint/types/time"
)

var (
	stakeDenom = "stake"
	feeDenom   = "fee"
)

func initBaseAccount() (*authtypes.BaseAccount, sdk.Coins) {
	_, _, addr := testdata.KeyTestPubAddr()
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}

func TestComputeClawback(t *testing.T) {
	c := sdk.NewCoins
	fee := func(x int64) sdk.Coin { return sdk.NewInt64Coin(feeDenom, x) }
	stake := func(x int64) sdk.Coin { return sdk.NewInt64Coin(stakeDenom, x) }
	now := tmtime.Now()
	lockupPeriods := Periods{
		{Length: int64(12 * 3600), Amount: c(fee(1000), stake(100))}, // noon
	}
	vestingPeriods := Periods{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
		{Length: int64(6 * 3600), Amount: c(fee(200), stake(50))}, // 3pm
		{Length: int64(2 * 3600), Amount: c(fee(200))},            // 5pm
		{Length: int64(1 * 3600), Amount: c(fee(200))},            // 6pm
	}

	bacc, origCoins := initBaseAccount()
	va := NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)

	amt := va.computeClawback(now.Unix())
	require.Equal(t, c(fee(1000), stake(100)), amt)
	require.Equal(t, c(), va.OriginalVesting)
	require.Equal(t, 0, len(va.LockupPeriods))
	require.Equal(t, 0, len(va.VestingPeriods))

	va2 := NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	amt = va2.computeClawback(now.Add(11 * time.Hour).Unix())
	require.Equal(t, c(fee(600), stake(50)), amt)
	require.Equal(t, c(fee(400), stake(50)), va2.OriginalVesting)
	require.Equal(t, []Period{{Length: int64(12 * 3600), Amount: c(fee(400), stake(50))}}, va2.LockupPeriods)
	require.Equal(t, []Period{
		{Length: int64(8 * 3600), Amount: c(fee(200))},            // 8am
		{Length: int64(1 * 3600), Amount: c(fee(200), stake(50))}, // 9am
	}, va2.VestingPeriods)

	va3 := NewClawbackVestingAccount(bacc, sdk.AccAddress([]byte("funder")), origCoins, now.Unix(), lockupPeriods, vestingPeriods)
	amt = va3.computeClawback(now.Add(23 * time.Hour).Unix())
	require.Equal(t, c(), amt)
}
