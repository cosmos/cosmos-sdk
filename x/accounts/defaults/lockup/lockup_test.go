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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func setup(t *testing.T, ctx context.Context, ss store.KVStoreService) *BaseLockup {
	t.Helper()
	deps := makeMockDependencies(ss)
	owner := "owner"

	baseLockup := newBaseLockup(deps)
	_, err := baseLockup.Init(ctx, &lockuptypes.MsgInitLockupAccount{
		Owner:   owner,
		EndTime: time.Now().Add(time.Minute),
	})
	require.NoError(t, err)

	return baseLockup
}

func TestInitLockupAccount(t *testing.T) {
	ctx, ss := newMockContext(t)
	deps := makeMockDependencies(ss)
	owner := "owner"

	baseLockup := newBaseLockup(deps)

	testcases := []struct {
		name   string
		msg    lockuptypes.MsgInitLockupAccount
		expErr error
	}{
		{
			"successfully init",
			lockuptypes.MsgInitLockupAccount{
				Owner:   owner,
				EndTime: time.Now().Add(10 * time.Second),
			},
			nil,
		},
	}

	for _, test := range testcases {
		_, err := baseLockup.Init(ctx, &test.msg)
		if test.expErr != nil {
			require.Equal(t, test.expErr, err)
			continue
		}
		require.NoError(t, err)
	}
}

func TestTrackingDelegation(t *testing.T) {
	ctx, ss := newMockContext(t)

	mockBalances := sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10)))

	testcases := []struct {
		name                   string
		amt                    sdk.Coins
		lockedCoins            sdk.Coins
		expDelegatedLockingAmt sdk.Coins
		expDelegatedFreeAmt    sdk.Coins
		malaete                func(ctx context.Context, bv *BaseLockup)
		expErr                 error
	}{
		{
			"delegate amount less than the vesting amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(1))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(1))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(0))),
			nil,
			nil,
		},
		{
			"delegate amount less than the vesting amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(1))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(4))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(0))),
			func(ctx context.Context, bv *BaseLockup) {
				err := bv.DelegatedLocking.Set(ctx, "test", math.NewInt(3))
				require.NoError(t, err)
			},
			nil,
		},
		{
			"delegate amount partially exceeds the vesting amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(2))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(1))),
			func(ctx context.Context, bv *BaseLockup) {
				err := bv.DelegatedLocking.Set(ctx, "test", math.NewInt(4))
				require.NoError(t, err)
			},
			nil,
		},
		{
			"delegate amount exceeds the vesting amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(6))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			func(ctx context.Context, bv *BaseLockup) {
				err := bv.DelegatedLocking.Set(ctx, "test", math.NewInt(4))
				require.NoError(t, err)
			},
			nil,
		},
		{
			"balances less than amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(11))),
			nil,
			nil,
			nil,
			nil,
			sdkerrors.ErrInvalidCoins.Wrap("delegation attempt with zero coins for staking denom or insufficient funds"),
		},
		{
			"zero amount",
			sdk.Coins{sdk.NewCoin("test", math.ZeroInt())},
			nil,
			nil,
			nil,
			nil,
			sdkerrors.ErrInvalidCoins.Wrap("delegation attempt with zero coins for staking denom or insufficient funds"),
		},
	}

	for _, test := range testcases {
		baseLockup := setup(t, ctx, ss)

		if test.malaete != nil {
			test.malaete(ctx, baseLockup)
		}

		err := baseLockup.TrackDelegation(ctx, mockBalances, test.lockedCoins, test.amt)
		if test.expErr != nil {
			require.EqualError(t, err, test.expErr.Error(), test.name+" error not equal")
			continue
		}

		delegatedVesting, err := baseLockup.DelegatedLocking.Get(ctx, "test")
		require.NoError(t, err)
		delegatedFree, err := baseLockup.DelegatedFree.Get(ctx, "test")
		require.NoError(t, err)

		require.Equal(t, test.expDelegatedLockingAmt.AmountOf("test"), delegatedVesting, test.name+" delegated locking amount must be equal")
		require.Equal(t, test.expDelegatedFreeAmt.AmountOf("test"), delegatedFree, test.name+" delegated free amount must be equal")
	}
}

func TestTrackingUnDelegation(t *testing.T) {
	ctx, ss := newMockContext(t)

	testcases := []struct {
		name                   string
		amt                    sdk.Coins
		expDelegatedLockingAmt sdk.Coins
		expDelegatedFreeAmt    sdk.Coins
		expErr                 error
	}{
		{
			"undelegate amount less than delegated free amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(1))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(5))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(4))),
			nil,
		},
		{
			"undelegate amount partially exceed the delegated free amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(6))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(4))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(0))),
			nil,
		},
		{
			"undelegate amount exceed the delegated free amount",
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(10))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(0))),
			sdk.NewCoins(sdk.NewCoin("test", math.NewInt(0))),
			nil,
		},
		{
			"zero amount",
			sdk.Coins{sdk.NewCoin("test", math.NewInt(0))},
			nil,
			nil,
			sdkerrors.ErrInvalidCoins.Wrap("undelegation attempt with zero coins for staking denom"),
		},
	}

	for _, test := range testcases {
		baseLockup := setup(t, ctx, ss)
		err := baseLockup.DelegatedLocking.Set(ctx, "test", math.NewInt(5))
		require.NoError(t, err)
		err = baseLockup.DelegatedFree.Set(ctx, "test", math.NewInt(5))
		require.NoError(t, err)

		err = baseLockup.TrackUndelegation(ctx, test.amt)
		if test.expErr != nil {
			require.EqualError(t, err, test.expErr.Error(), test.name+" error not equal")
			continue
		}

		delegatedVesting, err := baseLockup.DelegatedLocking.Get(ctx, "test")
		require.NoError(t, err)
		delegatedFree, err := baseLockup.DelegatedFree.Get(ctx, "test")
		require.NoError(t, err)

		require.Equal(t, test.expDelegatedLockingAmt.AmountOf("test"), delegatedVesting, "delegated locking amount must be equal")
		require.Equal(t, test.expDelegatedFreeAmt.AmountOf("test"), delegatedFree, "delegated free amount must be equal")
	}
}

func TestGetNotBondedLockedCoin(t *testing.T) {
	ctx, ss := newMockContext(t)
	sdkCtx := sdk.NewContext(nil, true, log.NewNopLogger()).WithContext(ctx).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	testcases := []struct {
		name        string
		lockedCoin  sdk.Coin
		expLockCoin sdk.Coin
		malaete     func(ctx context.Context, bv *BaseLockup)
	}{
		{
			"locked amount less than delegated locking amount",
			sdk.NewCoin("test", math.NewInt(1)),
			sdk.NewCoin("test", math.NewInt(0)),
			func(ctx context.Context, bv *BaseLockup) {
				err := bv.DelegatedLocking.Set(ctx, "test", math.NewInt(5))
				require.NoError(t, err)
			},
		},
		{
			"locked amount more than delegated locking amount",
			sdk.NewCoin("test", math.NewInt(5)),
			sdk.NewCoin("test", math.NewInt(1)),
			func(ctx context.Context, bv *BaseLockup) {
				err := bv.DelegatedLocking.Set(ctx, "test", math.NewInt(4))
				require.NoError(t, err)
			},
		},
	}

	for _, test := range testcases {
		baseLockup := setup(t, sdkCtx, ss)
		test.malaete(sdkCtx, baseLockup)

		lockedCoin, err := baseLockup.GetNotBondedLockedCoin(sdkCtx, test.lockedCoin, "test")
		require.NoError(t, err)

		require.True(t, test.expLockCoin.Equal(lockedCoin), test.name+" locked amount must be equal")
	}
}

func TestQueryLockupAccountBaseInfo(t *testing.T) {
	ctx, ss := newMockContext(t)

	baseLockup := setup(t, ctx, ss)

	res, err := baseLockup.QueryLockupAccountBaseInfo(ctx, &lockuptypes.QueryLockupAccountInfoRequest{})
	require.Equal(t, res.OriginalLocking.AmountOf("test"), math.NewInt(10))
	require.Equal(t, res.Owner, "owner")
	require.NoError(t, err)
}
