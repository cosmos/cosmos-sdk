package vesting

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*ContinuousVestingAccount)(nil)
)

// NewContinuousVestingAccount creates a new ContinuousVestingAccount object.
func NewContinuousVestingAccount(d accountstd.Dependencies) (*ContinuousVestingAccount, error) {
	baseVesting := NewBaseVesting(d)

	continuousVestingAccount := ContinuousVestingAccount{
		BaseVesting: baseVesting,
		StartTime:   collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", sdk.IntValue),
	}

	return &continuousVestingAccount, nil
}

type ContinuousVestingAccount struct {
	*BaseVesting
	StartTime collections.Item[math.Int]
}

func (cva ContinuousVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	if msg.StartTime < 0 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	if msg.EndTime <= 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid end time")
	}

	if msg.EndTime <= msg.StartTime {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid start and end time (must be start < end)")
	}

	hs := cva.headerService.GetHeaderInfo(ctx)

	start := msg.StartTime
	if msg.StartTime == 0 {
		start = hs.Time.Unix()
	}

	err := cva.StartTime.Set(ctx, math.NewInt(start))
	if err != nil {
		return nil, err
	}

	return cva.BaseVesting.Init(ctx, msg)
}

func (cva *ContinuousVestingAccount) ExecuteMessages(ctx context.Context, msg *account_abstractionv1.MsgExecute) (
	*account_abstractionv1.MsgExecuteResponse, error,
) {
	return cva.BaseVesting.ExecuteMessages(ctx, msg, cva.GetVestingCoins)
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (cva ContinuousVestingAccount) GetVestCoinsInfo(ctx context.Context, blockTime time.Time) (vestedCoins, vestingCoins sdk.Coins, err error) {
	vestedCoins = sdk.Coins{}
	vestingCoins = sdk.Coins{}

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	endTime, err := cva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	var originalVesting sdk.Coins
	cva.IterateCoinEntries(ctx, cva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	if startTime.GTE(math.NewInt(blockTime.Unix())) {
		return vestedCoins, originalVesting, nil
	} else if endTime.LTE(math.NewInt(blockTime.Unix())) {
		return originalVesting, vestingCoins, nil
	}

	// calculate the vesting scalar
	x := math.NewInt(blockTime.Unix()).Sub(startTime).Int64()
	y := endTime.Sub(startTime).Int64()
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	for _, ovc := range originalVesting {
		vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	vestingCoins = originalVesting.Sub(vestedCoins...)

	return vestedCoins, vestingCoins, nil
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	_, vestingCoins, err := cva.GetVestCoinsInfo(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return vestingCoins, nil
}

func (cva ContinuousVestingAccount) QueryVestCoinsInfo(ctx context.Context, msg *vestingtypes.QueryVestCoinsInfoRequest) (
	*vestingtypes.QueryVestCoinsInfoResponse, error,
) {
	hs := cva.headerService.GetHeaderInfo(ctx)
	vestedCoins, vestingCoins, err := cva.GetVestCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	return &vestingtypes.QueryVestCoinsInfoResponse{
		VestedVesting: vestedCoins,
		VestingCoins:  vestingCoins,
	}, nil
}

func (cva ContinuousVestingAccount) QueryStartTime(ctx context.Context, msg *vestingtypes.QueryStartTimeRequest) (
	*vestingtypes.QueryStartTimeResponse, error,
) {
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &vestingtypes.QueryStartTimeResponse{
		StartTime: startTime.Int64(),
	}, nil
}

// Implement smart account interface
func (cva ContinuousVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, cva.Init)
}

func (cva ContinuousVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, cva.ExecuteMessages)
	cva.BaseVesting.RegisterExecuteHandlers(builder)
}

func (cva ContinuousVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, cva.QueryStartTime)
	accountstd.RegisterQueryHandler(builder, cva.QueryVestCoinsInfo)
	cva.BaseVesting.RegisterQueryHandlers(builder)
}
