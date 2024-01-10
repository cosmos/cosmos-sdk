package vesting

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"
	"github.com/cosmos/cosmos-sdk/codec"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	to := accountstd.Whoami(ctx)
	if to == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("Cannot find account address from context")
	}

	toAddress, err := pva.AddressCodec.BytesToString(to)
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
		pva.VestingPeriods.Set(ctx, fmt.Sprint(i), period)
	}

	sortedAmt := totalCoins.Sort()
	for _, coin := range sortedAmt {
		pva.OriginalVesting.Set(ctx, coin.Denom, coin.Amount)
	}

	pva.StartTime.Set(ctx, math.NewInt(msg.StartTime))
	pva.EndTime.Set(ctx, math.NewInt(endTime))

	// Send token to new vesting account
	sendMsg := banktypes.NewMsgSend(msg.FromAddress, toAddress, totalCoins)
	anyMsg, err := codectypes.NewAnyWithValue(sendMsg)
	if err != nil {
		return nil, err
	}

	if _, err = accountstd.ExecModuleAnys(ctx, []*codectypes.Any{anyMsg}); err != nil {
		return nil, err
	}

	return &vestingtypes.MsgInitPeriodicVestingAccountResponse{}, nil
}

// --------------- execute -----------------

func (pva *PeriodicVestingAccount) ExecuteMessages(ctx context.Context, msg *vestingtypes.MsgExecuteMessages) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	return pva.BaseVestingAccount.ExecuteMessages(ctx, msg, pva.GetVestingCoins)
}

// ----------------- Query --------------------

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (bva BaseVestingAccount) IteratePeriods(
	ctx context.Context,
	entries collections.Map[string, vestingtypes.Period],
	cb func(_ string, value vestingtypes.Period) bool,
) {
	err := entries.Walk(ctx, nil, func(key string, value vestingtypes.Period) (stop bool, err error) {
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
	pva.IterateEntries(ctx, pva.OriginalVesting, func(key string, value math.Int) (stop bool) {
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

	pva.IteratePeriods(ctx, pva.VestingPeriods, func(_ string, period vestingtypes.Period) (stop bool) {
		x := math.NewInt(blockTime.Unix()).Sub(currentPeriodStartTime).Int64()
		if x < period.Length {
			return true
		}

		vestedCoins = vestedCoins.Add(period.Amount...)

		// update the start time of the next period
		pva.StartTime.Set(ctx, currentPeriodStartTime.Add(math.NewInt(period.Length)))
		return false
	})

	return vestedCoins, nil
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	var originalVesting sdk.Coins
	pva.IterateEntries(ctx, pva.OriginalVesting, func(key string, value math.Int) (stop bool) {
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
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestedCoins, err := pva.GetVestedCoins(ctx, sdkctx.HeaderInfo().Time)
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
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)
	vestingCoins, err := pva.GetVestingCoins(ctx, sdkctx.BlockHeader().Time)
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

// Implement smart account interface
func (pva PeriodicVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, pva.Init)
}

func (pva PeriodicVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, pva.ExecuteMessages)
}

func (pva PeriodicVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, pva.QueryStartTime)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, pva.QueryVestingCoins)
	pva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
