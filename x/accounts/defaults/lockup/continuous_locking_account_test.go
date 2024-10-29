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

func setupContinousAccount(t *testing.T, ctx context.Context, ss store.KVStoreService) *ContinuousLockingAccount {
	t.Helper()
	deps := makeMockDependencies(ss)
	owner := "owner" //nolint:goconst // adding constants for this would impede readability

	acc, err := NewContinuousLockingAccount(deps)
	require.NoError(t, err)
	_, err = acc.Init(ctx, &lockuptypes.MsgInitLockupAccount{
		Owner:     owner,
		EndTime:   time.Now().Add(time.Minute * 2),
		StartTime: time.Now(),
	})
	require.NoError(t, err)

	return acc
}

func TestContinousAccountDelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupContinousAccount(t, sdkCtx, ss)
	_, err := acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err := acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked half of the original locking amount
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: startTime.Add(time.Minute * 1),
	})

	_, err = acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(5)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(5)))

	delFree, err := acc.DelegatedFree.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delFree.Equal(math.NewInt(1)))
}

func TestContinousAccountUndelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupContinousAccount(t, sdkCtx, ss)
	// Delegate first
	_, err := acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err := acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))

	// Undelegate
	_, err = acc.Undelegate(sdkCtx, &lockuptypes.MsgUndelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.ZeroInt()))

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked half of the original locking amount
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: startTime.Add(time.Minute * 1),
	})

	_, err = acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(6)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(5)))

	delFree, err := acc.DelegatedFree.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delFree.Equal(math.NewInt(1)))

	// Undelegate
	_, err = acc.Undelegate(sdkCtx, &lockuptypes.MsgUndelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(4)),
	})
	require.NoError(t, err)

	delLocking, err = acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(2)))

	delFree, err = acc.DelegatedFree.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delFree.Equal(math.ZeroInt()))
}

func TestContinousAccountSendCoins(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupContinousAccount(t, sdkCtx, ss)
	_, err := acc.SendCoins(sdkCtx, &lockuptypes.MsgSend{
		Sender:    "owner",
		ToAddress: "receiver",
		Amount:    sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
	})
	require.Error(t, err)

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked half of the original locking amount
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

func TestContinousAccountWithdrawUnlockedCoins(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupContinousAccount(t, sdkCtx, ss)
	_, err := acc.WithdrawUnlockedCoins(sdkCtx, &lockuptypes.MsgWithdraw{
		Withdrawer: "owner",
		ToAddress:  "receiver",
		Denoms:     []string{"test"},
	})
	require.Error(t, err)

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// Update context time to unlocked half of the original locking amount
	sdkCtx = sdkCtx.WithHeaderInfo(header.Info{
		Time: startTime.Add(time.Minute * 1),
	})

	_, err = acc.WithdrawUnlockedCoins(sdkCtx, &lockuptypes.MsgWithdraw{
		Withdrawer: "owner",
		ToAddress:  "receiver",
		Denoms:     []string{"test"},
	})
	require.NoError(t, err)
}

func TestContinousAccountGetLockCoinInfo(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupContinousAccount(t, sdkCtx, ss)

	unlocked, locked, err := acc.GetLockCoinsInfo(sdkCtx, time.Now())
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.ZeroInt()))
	require.True(t, locked.AmountOf("test").Equal(math.NewInt(10)))

	startTime, err := acc.StartTime.Get(sdkCtx)
	require.NoError(t, err)

	// unlocked half locked token
	unlocked, locked, err = acc.GetLockCoinsInfo(sdkCtx, startTime.Add(time.Minute*1))
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.NewInt(5)))
	require.True(t, locked.AmountOf("test").Equal(math.NewInt(5)))

	// unlocked full locked token
	unlocked, locked, err = acc.GetLockCoinsInfo(sdkCtx, startTime.Add(time.Minute*2))
	require.NoError(t, err)
	require.True(t, unlocked.AmountOf("test").Equal(math.NewInt(10)))
	require.True(t, locked.AmountOf("test").Equal(math.ZeroInt()))
}
