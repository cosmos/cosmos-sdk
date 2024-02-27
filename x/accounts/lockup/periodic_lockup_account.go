package vesting

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/lockup/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*PeriodicVestingAccount)(nil)
)

// NewPeriodicVestingAccount creates a new PeriodicVestingAccount object.
func NewPeriodicVestingAccount(d accountstd.Dependencies) (*PeriodicVestingAccount, error) {
	baseVesting := NewBaseVesting(d)

	periodicsVestingAccount := PeriodicVestingAccount{
		BaseVesting:    baseVesting,
		StartTime:      collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
		VestingPeriods: collections.NewVec(d.SchemaBuilder, VestingPeriodsPrefix, "vesting_periods", codec.CollValue[vestingtypes.Period](d.LegacyStateCodec)),
	}

	return &periodicsVestingAccount, nil
}

type PeriodicVestingAccount struct {
	*BaseVesting
	StartTime      collections.Item[time.Time]
	VestingPeriods collections.Vec[vestingtypes.Period]
}

func (pva PeriodicVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitPeriodicVestingAccount) (*vestingtypes.MsgInitPeriodicVestingAccountResponse, error) {
	owner, err := pva.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}

	if msg.StartTime.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid start time of %s", msg.StartTime.String())
	}

	totalCoins := sdk.Coins{}
	endTime := msg.StartTime
	for _, period := range msg.VestingPeriods {
		if period.Length.Seconds() <= 0 {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period duration length %d", period.Length)
		}

		if err := validateAmount(period.Amount); err != nil {
			return nil, err
		}

		totalCoins = totalCoins.Add(period.Amount...)
		// Calculate end time
		endTime = endTime.Add(period.Length)
		err = pva.VestingPeriods.Push(ctx, period)
		if err != nil {
			return nil, err
		}
	}

	sortedAmt := totalCoins.Sort()
	for _, coin := range sortedAmt {
		err := pva.OriginalVesting.Set(ctx, coin.Denom, coin.Amount)
		if err != nil {
			return nil, err
		}
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

	return &vestingtypes.MsgInitPeriodicVestingAccountResponse{}, nil
}

func (pva *PeriodicVestingAccount) ExecuteMessages(ctx context.Context, msg *vestingtypes.MsgExecuteMessages) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	return pva.BaseVesting.ExecuteMessages(ctx, msg, pva.GetVestingCoinWithDenoms)
}

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (pva PeriodicVestingAccount) IteratePeriods(
	ctx context.Context,
	cb func(value vestingtypes.Period) (bool, error),
) error {
	err := pva.VestingPeriods.Walk(ctx, nil, func(_ uint64, value vestingtypes.Period) (stop bool, err error) {
		return cb(value)
	})
	if err != nil {
		return err
	}

	return nil
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva PeriodicVestingAccount) GetVestCoinsInfo(ctx context.Context, blockTime time.Time) (vestedCoins, vestingCoins sdk.Coins, err error) {
	vestedCoins = sdk.Coins{}
	vestingCoins = sdk.Coins{}

	// We must handle the case where the start time for a vesting account has
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
	originalVesting := sdk.Coins{}
	err = pva.IterateCoinEntries(ctx, pva.OriginalVesting, func(key string, value math.Int) (stop bool, err error) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if blockTime.Before(startTime) {
		return vestedCoins, originalVesting, nil
	} else if blockTime.After(endTime) {
		return originalVesting, vestingCoins, nil
	}

	// track the start time of the next period
	currentPeriodStartTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = pva.IteratePeriods(ctx, func(period vestingtypes.Period) (stop bool, err error) {
		x := blockTime.Sub(currentPeriodStartTime)
		if x.Seconds() < period.Length.Seconds() {
			return true, nil
		}

		vestedCoins = vestedCoins.Add(period.Amount...)

		// update the start time of the next period
		err = pva.StartTime.Set(ctx, currentPeriodStartTime.Add(period.Length))
		if err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	vestingCoins = originalVesting.Sub(vestedCoins...)

	return vestedCoins, vestingCoins, err
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	_, vestingCoins, err := pva.GetVestCoinsInfo(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return vestingCoins, nil
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva PeriodicVestingAccount) GetVestCoinInfoWithDenom(ctx context.Context, blockTime time.Time, denom string) (vestedCoin, vestingCoin *sdk.Coin, err error) {
	// We must handle the case where the start time for a vesting account has
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
	originalVestingAmt, err := pva.OriginalVesting.Get(ctx, denom)
	if err != nil {
		return nil, nil, err
	}

	originalVesting := sdk.NewCoin(denom, originalVestingAmt)
	if err != nil {
		return nil, nil, err
	}
	if blockTime.Before(startTime) {
		return &sdk.Coin{}, &originalVesting, nil
	} else if blockTime.After(endTime) {
		return &originalVesting, &sdk.Coin{}, nil
	}

	// track the start time of the next period
	currentPeriodStartTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	vested := sdk.NewCoin(denom, math.ZeroInt())
	err = pva.IteratePeriods(ctx, func(period vestingtypes.Period) (stop bool, err error) {
		x := blockTime.Sub(currentPeriodStartTime)
		if x.Seconds() < period.Length.Seconds() {
			return true, nil
		}

		vested = vested.Add(sdk.NewCoin(denom, period.Amount.AmountOf(denom)))

		// update the start time of the next period
		err = pva.StartTime.Set(ctx, currentPeriodStartTime.Add(period.Length))
		if err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	vesting := originalVesting.Sub(vested)

	return &vested, &vesting, err
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoinWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	vestingCoins := sdk.Coins{}
	for _, denom := range denoms {
		_, vestingCoin, err := pva.GetVestCoinInfoWithDenom(ctx, blockTime, denom)
		if err != nil {
			return nil, err
		}
		vestingCoins = append(vestingCoins, *vestingCoin)
	}
	return vestingCoins, nil
}

func (pva PeriodicVestingAccount) QueryVestingAccountInfo(ctx context.Context, req *vestingtypes.QueryVestingAccountInfoRequest) (
	*vestingtypes.QueryVestingAccountInfoResponse, error,
) {
	resp, err := pva.BaseVesting.QueryVestingAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	hs := pva.headerService.GetHeaderInfo(ctx)
	vestedCoins, vestingCoins, err := pva.GetVestCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.VestedVesting = sdk.Coins{}
	resp.StartTime = &startTime
	resp.VestingCoins = vestingCoins
	resp.VestedVesting = vestedCoins
	return resp, nil
}

func (pva PeriodicVestingAccount) QueryVestingPeriods(ctx context.Context, msg *vestingtypes.QueryVestingPeriodsRequest) (
	*vestingtypes.QueryVestingPeriodsResponse, error,
) {
	vestingPeriods := []*vestingtypes.Period{}
	err := pva.IteratePeriods(ctx, func(period vestingtypes.Period) (stop bool, err error) {
		vestingPeriods = append(vestingPeriods, &period)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return &vestingtypes.QueryVestingPeriodsResponse{
		VestingPeriods: vestingPeriods,
	}, nil
}

// Implement smart account interface
func (pva PeriodicVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, pva.Init)
}

func (pva PeriodicVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, pva.ExecuteMessages)
}

func (pva PeriodicVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, pva.QueryVestingAccountInfo)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestingPeriods)
}
