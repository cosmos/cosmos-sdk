package vesting

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*PeriodicVestingAccount)(nil)
)

// Periodic Vesting Account

// NewPeriodicVestingAccount creates a new PeriodicVestingAccount object.
func NewPeriodicVestingAccount(d accountstd.Dependencies) (*PeriodicVestingAccount, error) {
	baseVestingAccount, err := NewBaseVestingAccount(d)

	periodicsVestingAccount := PeriodicVestingAccount{
		BaseVestingAccount: baseVestingAccount,
		StartTime:          collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", sdk.IntValue),
		VestingPeriods:     collections.NewMap(d.SchemaBuilder, VestingPeriodsPrefix, "vesting_periods", collections.StringKey, codec.CollValue[vestingtypes.Period](d.BinaryCodec)),
	}

	return &periodicsVestingAccount, err
}

type PeriodicVestingAccount struct {
	*BaseVestingAccount
	StartTime      collections.Item[math.Int]
	VestingPeriods collections.Map[string, vestingtypes.Period]
}

// --------------- Init -----------------

func (pva PeriodicVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitPeriodicVestingAccount) (*vestingtypes.MsgInitPeriodicVestingAccountResponse, error) {
	sender := accountstd.Sender(ctx)
	if sender == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find sender address from context")
	}
	to := accountstd.Whoami(ctx)
	if to == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find account address from context")
	}

	toAddress, err := pva.addressCodec.BytesToString(to)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}
	fromAddress, err := pva.addressCodec.BytesToString(sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	if msg.StartTime < 1 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	var totalCoins sdk.Coins
	endTime := msg.StartTime
	for i, period := range msg.VestingPeriods {
		if period.Length < 1 {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid period length of %d in period %d, length must be greater than 0", period.Length, i)
		}

		if err := validateAmount(period.Amount); err != nil {
			return nil, err
		}

		totalCoins = totalCoins.Add(period.Amount...)
		// Calculate end time
		endTime += period.Length
		err = pva.VestingPeriods.Set(ctx, fmt.Sprint(i), period)
		if err != nil {
			return nil, err
		}
	}

	sortedAmt := totalCoins.Sort()
	for _, coin := range sortedAmt {
		pva.OriginalVesting.Set(ctx, coin.Denom, coin.Amount)
	}

	err = pva.StartTime.Set(ctx, math.NewInt(msg.StartTime))
	if err != nil {
		return nil, err
	}
	err = pva.EndTime.Set(ctx, math.NewInt(endTime))
	if err != nil {
		return nil, err
	}
	err = pva.Owner.Set(ctx, sender)
	if err != nil {
		return nil, err
	}

	// Send token to new vesting account
	sendMsg := banktypes.NewMsgSend(fromAddress, toAddress, totalCoins)

	if _, err = accountstd.ExecModule[banktypes.MsgSendResponse](ctx, sendMsg); err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitPeriodicVestingAccountResponse{}, nil
}

// --------------- execute -----------------

func (pva *PeriodicVestingAccount) ExecuteMessages(ctx context.Context, msg *account_abstractionv1.MsgExecute) (
	*account_abstractionv1.MsgExecuteResponse, error,
) {
	return pva.BaseVestingAccount.ExecuteMessages(ctx, msg, pva.GetVestingCoins)
}

// ----------------- Query --------------------

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (pva PeriodicVestingAccount) IteratePeriods(
	ctx context.Context,
	cb func(_ string, value vestingtypes.Period) bool,
) {
	err := pva.VestingPeriods.Walk(ctx, nil, func(key string, value vestingtypes.Period) (stop bool, err error) {
		return cb(key, value), nil
	})
	if err != nil {
		panic(err)
	}
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva PeriodicVestingAccount) GetVestedCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	endTime, err := pva.EndTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	var originalVesting sdk.Coins
	pva.IterateCoinEntries(ctx, pva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	if math.NewInt(blockTime.Unix()).LTE(startTime) {
		return vestedCoins, nil
	} else if math.NewInt(blockTime.Unix()).GTE(endTime) {
		return originalVesting, nil
	}

	// track the start time of the next period
	currentPeriodStartTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}

	pva.IteratePeriods(ctx, func(_ string, period vestingtypes.Period) (stop bool) {
		x := math.NewInt(blockTime.Unix()).Sub(currentPeriodStartTime).Int64()
		if x < period.Length {
			return true
		}

		vestedCoins = vestedCoins.Add(period.Amount...)

		// update the start time of the next period
		err = pva.StartTime.Set(ctx, currentPeriodStartTime.Add(math.NewInt(period.Length)))
		if err != nil {
			return true
		}
		return false
	})

	return vestedCoins, err
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	var originalVesting sdk.Coins
	pva.IterateCoinEntries(ctx, pva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	vestedCoins, err := pva.GetVestedCoins(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return originalVesting.Sub(vestedCoins...), nil
}

func (pva PeriodicVestingAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	hs := pva.headerService.GetHeaderInfo(ctx)
	vestedCoins, err := pva.GetVestedCoins(ctx, hs.Time)
	if err != nil {
		return nil, err
	}

	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (pva PeriodicVestingAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	hs := pva.headerService.GetHeaderInfo(ctx)
	vestingCoins, err := pva.GetVestingCoins(ctx, hs.Time)
	if err != nil {
		return nil, err
	}

	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
	}, nil
}

func (pva PeriodicVestingAccount) QueryStartTime(ctx context.Context, msg *vestingtypes.QueryStartTimeRequest) (
	*vestingtypes.QueryStartTimeResponse, error,
) {
	startTime, err := pva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &vestingtypes.QueryStartTimeResponse{
		StartTime: startTime.Int64(),
	}, nil
}

func (pva PeriodicVestingAccount) QueryVestingPeriods(ctx context.Context, msg *vestingtypes.QueryVestingPeriodsRequest) (
	*vestingtypes.QueryVestingPeriodsResponse, error,
) {
	var vestingPeriods []vestingtypes.Period
	pva.IteratePeriods(ctx, func(_ string, period vestingtypes.Period) (stop bool) {
		vestingPeriods = append(vestingPeriods, period)
		return false
	})
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
	pva.BaseVestingAccount.RegisterExecuteHandlers(builder)
}

func (pva PeriodicVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, pva.QueryStartTime)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestingCoins)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestingPeriods)
	pva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
