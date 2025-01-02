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

func setupDelayedAccount(t *testing.T, ctx context.Context, ss store.KVStoreService) *DelayedLockingAccount {
	t.Helper()
	deps := makeMockDependencies(ss)
	owner := "owner"

	acc, err := NewDelayedLockingAccount(deps)
	require.NoError(t, err)
	_, err = acc.Init(ctx, &lockuptypes.MsgInitLockupAccount{
		Owner:   owner,
		EndTime: time.Now().Add(time.Minute * 2),
	})
	require.NoError(t, err)

	return acc
}

func TestDelayedAccountDelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupDelayedAccount(t, sdkCtx, ss)
	_, err := acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err := acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))

	endTime, err := acc.EndTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked all the original locking amount
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: endTime.Add(time.Second),
	})

	_, err = acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: valAddress,
		Amount:           sdk.NewCoin("test", math.NewInt(5)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))

	delFree, err := acc.DelegatedFree.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delFree.Equal(math.NewInt(5)))
}

func TestDelayedAccountUndelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupDelayedAccount(t, sdkCtx, ss)
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

func TestDelayedAccountSendCoins(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupDelayedAccount(t, sdkCtx, ss)
	_, err := acc.SendCoins(sdkCtx, &lockuptypes.MsgSend{
		Sender:    "owner",
		ToAddress: "receiver",
		Amount:    sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
	})
	require.Error(t, err)

	endTime, err := acc.EndTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked all the original locking amount
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: endTime.Add(time.Second),
	})

	_, err = acc.SendCoins(sdkCtx, &lockuptypes.MsgSend{
		Sender:    "owner",
		ToAddress: "receiver",
		Amount:    sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
	})
	require.NoError(t, err)
}

func TestDelayedAccountGetLockCoinInfo(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupDelayedAccount(t, sdkCtx, ss)

	unlocked, locked, err := acc.GetLockCoinsInfo(sdkCtx, time.Now())
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.ZeroInt()))
	require.True(t, locked.AmountOf("test").Equal(math.NewInt(10)))

	endTime, err := acc.EndTime.Get(sdkCtx)
	require.NoError(t, err)

	// unlocked full locked token
	unlocked, locked, err = acc.GetLockCoinsInfo(sdkCtx, endTime.Add(time.Second*1))
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.NewInt(10)))
	require.True(t, locked.AmountOf("test").Equal(math.ZeroInt()))
}
