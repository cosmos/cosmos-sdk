package lockup

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var _ accountstd.Interface = (*PeriodicLockingAccount)(nil)

// NewPeriodicLockingAccount creates a new PeriodicLockingAccount object.
func NewPeriodicLockingAccount(clawbackEnable bool) accountstd.AccountCreatorFunc {
	return func(d accountstd.Dependencies) (string, accountstd.Interface, error) {
		if clawbackEnable {
			baseClawback := newBaseClawback(d)

			return types.PERIODIC_LOCKING_ACCOUNT + types.CLAWBACK_ENABLE_PREFIX, PeriodicLockingAccount{
				BaseAccount:    baseClawback,
				StartTime:      collections.NewItem(d.SchemaBuilder, types.StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
				LockingPeriods: collections.NewVec(d.SchemaBuilder, types.LockingPeriodsPrefix, "locking_periods", codec.CollValue[types.Period](d.LegacyStateCodec))}, nil
		}

		baseLockup := newBaseLockup(d)
		return types.PERIODIC_LOCKING_ACCOUNT, PeriodicLockingAccount{
			BaseAccount:    baseLockup,
			StartTime:      collections.NewItem(d.SchemaBuilder, types.StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
			LockingPeriods: collections.NewVec(d.SchemaBuilder, types.LockingPeriodsPrefix, "locking_periods", codec.CollValue[types.Period](d.LegacyStateCodec))}, nil
	}
}

type PeriodicLockingAccount struct {
	types.BaseAccount
	StartTime      collections.Item[time.Time]
	LockingPeriods collections.Vec[types.Period]
}

func (pva PeriodicLockingAccount) Init(ctx context.Context, msg *types.MsgInitPeriodicLockingAccount) (*types.MsgInitPeriodicLockingAccountResponse, error) {
	hs := pva.GetHeaderService().HeaderInfo(ctx)

	if msg.StartTime.Before(hs.Time) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("start time %s should be after block time")
	}

	totalCoins := sdk.Coins{}
	endTime := msg.StartTime
	for _, period := range msg.LockingPeriods {
		if period.Length.Seconds() <= 0 {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period duration length %d", period.Length)
		}

		if period.Amount.IsZero() {
			return nil, sdkerrors.ErrInvalidCoins.Wrap("period amount cannot be zero")
		}

		totalCoins = totalCoins.Add(period.Amount...)
		// Calculate end time
		endTime = endTime.Add(period.Length)
		err := pva.LockingPeriods.Push(ctx, period)
		if err != nil {
			return nil, err
		}
	}

	funds := accountstd.Funds(ctx)
	if !funds.Equal(totalCoins) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid funding amount, should be equal to total coins lockup")
	}

	sortedAmt := totalCoins.Sort()

	// TODO: maybe add periods in MsgInitLockupAccount so periodic account can use it
	msgInit := &types.MsgInitLockupAccount{
		Owner:     msg.Owner,
		EndTime:   endTime,
		StartTime: msg.StartTime,
		Admin:     msg.Admin,
	}

	_, err := pva.BaseAccount.Init(ctx, msgInit, sortedAmt)
	if err != nil {
		return nil, err
	}

	err = pva.StartTime.Set(ctx, msg.StartTime)
	if err != nil {
		return nil, err
	}

	return &types.MsgInitPeriodicLockingAccountResponse{}, nil
}

func (pva *PeriodicLockingAccount) Delegate(ctx context.Context, msg *types.MsgDelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := pva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not delegate enable")
	}
	return baseLockup.Delegate(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

func (pva *PeriodicLockingAccount) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := pva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not undelegate enable")
	}
	return baseLockup.Undelegate(ctx, msg)
}

func (pva *PeriodicLockingAccount) SendCoins(ctx context.Context, msg *types.MsgSend) (
	*types.MsgExecuteMessagesResponse, error,
) {
	return pva.BaseAccount.SendCoins(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

func (pva *PeriodicLockingAccount) WithdrawUnlockedCoins(ctx context.Context, msg *types.MsgWithdraw) (
	*types.MsgWithdrawResponse, error,
) {
	return pva.BaseAccount.WithdrawUnlockedCoins(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

func (pva *PeriodicLockingAccount) ClawbackFunds(ctx context.Context, msg *types.MsgClawback) (
	*types.MsgClawbackResponse, error,
) {
	baseClawback, ok := pva.BaseAccount.(*BaseClawback)
	if !ok {
		return nil, fmt.Errorf("clawback is not enable for this account type")
	}
	return baseClawback.ClawbackFunds(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

// IteratePeriods iterates over all the Period entries.
func (pva PeriodicLockingAccount) IteratePeriods(
	ctx context.Context,
	cb func(value types.Period) (bool, error),
) error {
	err := pva.LockingPeriods.Walk(ctx, nil, func(_ uint64, value types.Period) (stop bool, err error) {
		return cb(value)
	})
	if err != nil {
		return err
	}

	return nil
}

// GetLockCoinsInfo returns the total number of locked and unlocked coins.
func (pva PeriodicLockingAccount) GetLockCoinsInfo(ctx context.Context, blockTime time.Time) (unlockedCoins, lockedCoins sdk.Coins, err error) {
	unlockedCoins = sdk.Coins{}
	lockedCoins = sdk.Coins{}

	// We must handle the case where the start time for a lockup account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	endTime, err := pva.GetEndTime().Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLocking := sdk.Coins{}
	err = IterateCoinEntries(ctx, pva.GetOriginalFunds(), func(key string, value math.Int) (stop bool, err error) {
		originalLocking = append(originalLocking, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if blockTime.Before(startTime) {
		return unlockedCoins, originalLocking, nil
	} else if blockTime.After(endTime) {
		return originalLocking, lockedCoins, nil
	}

	// track the start time of the next period
	currentPeriodStartTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = pva.IteratePeriods(ctx, func(period types.Period) (stop bool, err error) {
		x := blockTime.Sub(currentPeriodStartTime)
		if x.Seconds() < period.Length.Seconds() {
			return true, nil
		}

		unlockedCoins = unlockedCoins.Add(period.Amount...)

		// update the start time of the next period
		currentPeriodStartTime = currentPeriodStartTime.Add(period.Length)
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	lockedCoins = originalLocking.Sub(unlockedCoins...)

	return unlockedCoins, lockedCoins, err
}

// GetLockedCoins returns the total number of locked coins. If no coins are
// locked, nil is returned.
func (pva PeriodicLockingAccount) GetLockedCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	_, vestingCoins, err := pva.GetLockCoinsInfo(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return vestingCoins, nil
}

// GetLockCoinInfoWithDenom returns the total number of locked and unlocked coin for a specific denom.
func (pva PeriodicLockingAccount) GetLockCoinInfoWithDenom(ctx context.Context, blockTime time.Time, denom string) (unlockedCoin, lockedCoin *sdk.Coin, err error) {
	// We must handle the case where the start time for a lockup account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	endTime, err := pva.GetEndTime().Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLockingAmt, err := pva.GetOriginalFunds().Get(ctx, denom)
	if err != nil {
		return nil, nil, err
	}

	originalLockingCoin := sdk.NewCoin(denom, originalLockingAmt)

	if blockTime.Before(startTime) {
		return &sdk.Coin{}, &originalLockingCoin, nil
	} else if blockTime.After(endTime) {
		return &originalLockingCoin, &sdk.Coin{}, nil
	}

	// track the start time of the next period
	currentPeriodStartTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	unlocked := sdk.NewCoin(denom, math.ZeroInt())
	err = pva.IteratePeriods(ctx, func(period types.Period) (stop bool, err error) {
		x := blockTime.Sub(currentPeriodStartTime)
		if x.Seconds() < period.Length.Seconds() {
			return true, nil
		}

		unlocked = unlocked.Add(sdk.NewCoin(denom, period.Amount.AmountOf(denom)))

		// update the start time of the next period
		currentPeriodStartTime = currentPeriodStartTime.Add(period.Length)
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	locked := originalLockingCoin.Sub(unlocked)

	return &unlocked, &locked, err
}

// GetLockedCoinsWithDenoms returns the total number of locked coins. If no coins are
// locked, nil is returned.
func (pva PeriodicLockingAccount) GetLockedCoinsWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	lockedCoins := sdk.Coins{}
	for _, denom := range denoms {
		_, lockedCoin, err := pva.GetLockCoinInfoWithDenom(ctx, blockTime, denom)
		if err != nil {
			return nil, err
		}
		lockedCoins = append(lockedCoins, *lockedCoin)
	}
	return lockedCoins, nil
}

func (pva PeriodicLockingAccount) QueryLockupAccountInfo(ctx context.Context, req *types.QueryLockupAccountInfoRequest) (
	*types.QueryLockupAccountInfoResponse, error,
) {
	resp, err := pva.BaseAccount.QueryAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	hs := pva.GetHeaderService().HeaderInfo(ctx)
	unlockedCoins, lockedCoins, err := pva.GetLockCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.StartTime = &startTime
	resp.LockedCoins = lockedCoins
	resp.UnlockedCoins = unlockedCoins
	return resp, nil
}

func (pva PeriodicLockingAccount) QueryLockingPeriods(ctx context.Context, msg *types.QueryLockingPeriodsRequest) (
	*types.QueryLockingPeriodsResponse, error,
) {
	lockingPeriods := []*types.Period{}
	err := pva.IteratePeriods(ctx, func(period types.Period) (stop bool, err error) {
		lockingPeriods = append(lockingPeriods, &period)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return &types.QueryLockingPeriodsResponse{
		LockingPeriods: lockingPeriods,
	}, nil
}

// Implement smart account interface
func (pva PeriodicLockingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, pva.Init)
}

func (pva PeriodicLockingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, pva.Delegate)
	accountstd.RegisterExecuteHandler(builder, pva.SendCoins)
	accountstd.RegisterExecuteHandler(builder, pva.WithdrawUnlockedCoins)
	accountstd.RegisterExecuteHandler(builder, pva.ClawbackFunds)

	baseLockup, ok := pva.BaseAccount.(*BaseLockup)
	if ok {
		baseLockup.RegisterExecuteHandlers(builder)
	}
}

func (pva PeriodicLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, pva.QueryLockupAccountInfo)
	accountstd.RegisterQueryHandler(builder, pva.QueryLockingPeriods)
}
