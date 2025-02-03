package lockup

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	lockuptypes "cosmossdk.io/x/accounts/defaults/lockup/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupPeriodicAccount(t *testing.T, ctx context.Context, ss store.KVStoreService) *PeriodicLockingAccount {
	t.Helper()
	deps := makeMockDependencies(ss)
	owner := "owner"

	acc, err := NewPeriodicLockingAccount(deps)
	require.NoError(t, err)
	_, err = acc.Init(ctx, &lockuptypes.MsgInitPeriodicLockingAccount{
		Owner:     owner,
		StartTime: time.Now(),
		LockingPeriods: []lockuptypes.Period{
			{
				Amount: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
				Length: time.Minute,
			},
			{
				Amount: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(2))),
				Length: time.Minute,
			},
			{
				Amount: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(3))),
				Length: time.Minute,
			},
		},
	})
	require.NoError(t, err)

	return acc
}

func TestPeriodicAccountDelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPeriodicAccount(t, sdkCtx, ss)
	_, err := acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err := acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked first period token
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: startTime.Add(time.Minute * 1),
	})

	_, err = acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(5)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(5)))

	delFree, err := acc.DelegatedFree.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delFree.Equal(math.NewInt(1)))

	// Update context time to unlocked all token
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: startTime.Add(time.Minute * 3),
	})

	_, err = acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(4)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(5)))

	delFree, err = acc.DelegatedFree.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delFree.Equal(math.NewInt(5)))
}

func TestPeriodicAccountUndelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPeriodicAccount(t, sdkCtx, ss)
	// Delegate first
	_, err := acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err := acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))

	// Undelegate
	_, err = acc.Undelegate(sdkCtx, &lockuptypes.MsgUndelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	// sequence should be the previous one
	entries, err := acc.UnbondEntries.Get(sdkCtx, valAddress)
	require.NoError(t, err)
	require.Len(t, entries.Entries, 1)
	require.True(t, entries.Entries[0].Amount.Amount.Equal(math.NewInt(1)))
	require.True(t, entries.Entries[0].ValidatorAddress == valAddress)

	err = acc.checkUnbondingEntriesMature(sdkCtx)
	require.NoError(t, err)

	_, err = acc.UnbondEntries.Get(sdkCtx, valAddress)
	require.Error(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.ZeroInt()))
}

func TestPeriodicAccountSendCoins(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPeriodicAccount(t, sdkCtx, ss)
	_, err := acc.SendCoins(sdkCtx, &lockuptypes.MsgSend{
		Sender:    "owner",
		ToAddress: "receiver",
		Amount:    sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
	})
	require.Error(t, err)

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked first period token
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: startTime.Add(time.Minute * 1),
	})

	_, err = acc.SendCoins(sdkCtx, &lockuptypes.MsgSend{
		Sender:    "owner",
		ToAddress: "receiver",
		Amount:    sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
	})
	require.NoError(t, err)
}

func TestPeriodicAccountGetLockCoinInfo(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPeriodicAccount(t, sdkCtx, ss)

	unlocked, locked, err := acc.GetLockCoinsInfo(sdkCtx, time.Now())
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.ZeroInt()))
	require.True(t, locked.AmountOf("test").Equal(math.NewInt(10)))

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// unlocked first period locked token
	unlocked, locked, err = acc.GetLockCoinsInfo(sdkCtx, startTime.Add(time.Minute*1))
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.NewInt(5)))
	require.True(t, locked.AmountOf("test").Equal(math.NewInt(5)))

	// unlocked second period locked token
	unlocked, locked, err = acc.GetLockCoinsInfo(sdkCtx, startTime.Add(time.Minute*2))
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.NewInt(7)))
	require.True(t, locked.AmountOf("test").Equal(math.NewInt(3)))

	// unlocked third period locked token
	unlocked, locked, err = acc.GetLockCoinsInfo(sdkCtx, startTime.Add(time.Minute*3))
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.NewInt(10)))
	require.True(t, locked.AmountOf("test").Equal(math.ZeroInt()))
}
