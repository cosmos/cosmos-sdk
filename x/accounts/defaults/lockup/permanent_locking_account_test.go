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
	lockuptypes "cosmossdk.io/x/accounts/defaults/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupPermanentAccount(t *testing.T, ctx context.Context, ss store.KVStoreService) *PermanentLockingAccount {
	t.Helper()
	deps := makeMockDependencies(ss)
	owner := "owner"

	acc, err := NewPermanentLockingAccount(deps)
	require.NoError(t, err)
	_, err = acc.Init(ctx, &lockuptypes.MsgInitLockupAccount{
		Owner: owner,
	})
	require.NoError(t, err)

	return acc
}

func TestPermanentAccountDelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPermanentAccount(t, sdkCtx, ss)
	_, err := acc.Delegate(sdkCtx, &lockuptypes.MsgDelegate{
		Sender:           "owner",
		ValidatorAddress: "val_address",
		Amount:           sdk.NewCoin("test", math.NewInt(1)),
	})
	require.NoError(t, err)

	delLocking, err := acc.DelegatedLocking.Get(ctx, "test")
	require.NoError(t, err)
	require.True(t, delLocking.Equal(math.NewInt(1)))
}

func TestPermanentAccountUndelegate(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPermanentAccount(t, sdkCtx, ss)
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
}

func TestPermanentAccountSendCoins(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	acc := setupPermanentAccount(t, sdkCtx, ss)
	_, err := acc.SendCoins(sdkCtx, &lockuptypes.MsgSend{
		Sender:    "owner",
		ToAddress: "receiver",
		Amount:    sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
	})
	require.Error(t, err)
}
