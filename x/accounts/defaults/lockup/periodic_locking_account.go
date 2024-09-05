package lockup

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	lockuptypes "cosmossdk.io/x/accounts/defaults/lockup/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var _ accountstd.Interface = (*PeriodicLockingAccount)(nil)

// NewPeriodicLockingAccount creates a new PeriodicLockingAccount object.
func NewPeriodicLockingAccount(d accountstd.Dependencies) (*PeriodicLockingAccount, error) {
	baseLockup := newBaseLockup(d)

	periodicsVestingAccount := PeriodicLockingAccount{
		BaseLockup:     baseLockup,
		StartTime:      collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
		LockingPeriods: collections.NewVec(d.SchemaBuilder, LockingPeriodsPrefix, "locking_periods", codec.CollValue[lockuptypes.Period](d.LegacyStateCodec)),
	}

	return &periodicsVestingAccount, nil
}

type PeriodicLockingAccount struct {
	*BaseLockup
	StartTime      collections.Item[time.Time]
	LockingPeriods collections.Vec[lockuptypes.Period]
}

func (pva PeriodicLockingAccount) Init(ctx context.Context, msg *lockuptypes.MsgInitPeriodicLockingAccount) (*lockuptypes.MsgInitPeriodicLockingAccountResponse, error) {
	owner, err := pva.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}

	hs := pva.headerService.HeaderInfo(ctx)

	if msg.StartTime.Before(hs.Time) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("start time %s should be after block time")
	}

	totalCoins := sdk.Coins{}
	endTime := msg.StartTime
	for _, period := range msg.LockingPeriods {
		if period.Length.Seconds() <= 0 {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period duration length %d", period.Length)
		}

		if err := validateAmount(period.Amount); err != nil {
			return nil, err
		}

		totalCoins = totalCoins.Add(period.Amount...)
		// Calculate end time
		endTime = endTime.Add(period.Length)
		err = pva.LockingPeriods.Push(ctx, period)
		if err != nil {
			return nil, err
		}
	}

	funds := accountstd.Funds(ctx)
	if !funds.Equal(totalCoins) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid funding amount, should be equal to total coins lockup")
	}

	sortedAmt := totalCoins.Sort()
	for _, coin := range sortedAmt {
		err := pva.OriginalLocking.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}

		// Set initial value for all withdrawed token
		err = pva.WithdrawedCoins.Set(ctx, coin.Denom, math.ZeroInt())
		if err != nil {
			return nil, err
		}
	}

	bondDenom, err := getStakingDenom(ctx)
	if err != nil {
		return nil, err
	}

	// Set initial value for all locked token
	err = pva.DelegatedFree.Set(ctx, bondDenom, math.ZeroInt())
	if err != nil {
		return nil, err
	}

	// Set initial value for all locked token
	err = pva.DelegatedLocking.Set(ctx, bondDenom, math.ZeroInt())
	if err != nil {
		return nil, err
	}

	err = pva.StartTime.Set(ctx, msg.StartTime)
	if err != nil {
		return nil, err
	}
	err = pva.EndTime.Set(ctx, endTime)
	if err != nil {
		return nil, err
	}
	err = pva.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}

	return &lockuptypes.MsgInitPeriodicLockingAccountResponse{}, nil
}

func (pva *PeriodicLockingAccount) Delegate(ctx context.Context, msg *lockuptypes.MsgDelegate) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return pva.BaseLockup.Delegate(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

func (pva *PeriodicLockingAccount) SendCoins(ctx context.Context, msg *lockuptypes.MsgSend) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return pva.BaseLockup.SendCoins(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

func (pva *PeriodicLockingAccount) WithdrawUnlockedCoins(ctx context.Context, msg *lockuptypes.MsgWithdraw) (
	*lockuptypes.MsgWithdrawResponse, error,
) {
	return pva.BaseLockup.WithdrawUnlockedCoins(ctx, msg, pva.GetLockedCoinsWithDenoms)
}

// IteratePeriods iterates over all the Periods entries.
func (pva PeriodicLockingAccount) IteratePeriods(
	ctx context.Context,
	cb func(value lockuptypes.Period) (bool, error),
) error {
	err := pva.LockingPeriods.Walk(ctx, nil, func(_ uint64, value lockuptypes.Period) (stop bool, err error) {
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
	endTime, err := pva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLocking := sdk.Coins{}
	err = pva.IterateCoinEntries(ctx, pva.OriginalLocking, func(key string, value math.Int) (stop bool, err error) {
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

	err = pva.IteratePeriods(ctx, func(period lockuptypes.Period) (stop bool, err error) {
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
	endTime, err := pva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLockingAmt, err := pva.OriginalLocking.Get(ctx, denom)
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
	err = pva.IteratePeriods(ctx, func(period lockuptypes.Period) (stop bool, err error) {
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

func (pva PeriodicLockingAccount) QueryLockupAccountInfo(ctx context.Context, req *lockuptypes.QueryLockupAccountInfoRequest) (
	*lockuptypes.QueryLockupAccountInfoResponse, error,
) {
	resp, err := pva.BaseLockup.QueryLockupAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	hs := pva.headerService.HeaderInfo(ctx)
	unlockedCoins, lockedCoins, err := pva.GetLockCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.StartTime = &startTime
	resp.LockedCoins = lockedCoins
	resp.UnlockedCoins = unlockedCoins
	return resp, nil
}

func (pva PeriodicLockingAccount) QueryLockingPeriods(ctx context.Context, msg *lockuptypes.QueryLockingPeriodsRequest) (
	*lockuptypes.QueryLockingPeriodsResponse, error,
) {
	lockingPeriods := []*lockuptypes.Period{}
	err := pva.IteratePeriods(ctx, func(period lockuptypes.Period) (stop bool, err error) {
		lockingPeriods = append(lockingPeriods, &period)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return &lockuptypes.QueryLockingPeriodsResponse{
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
	pva.BaseLockup.RegisterExecuteHandlers(builder)
}

func (pva PeriodicLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, pva.QueryLockupAccountInfo)
	accountstd.RegisterQueryHandler(builder, pva.QueryLockingPeriods)
}
