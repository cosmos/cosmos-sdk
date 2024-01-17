package vesting

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*ContinuousVestingAccount)(nil)
)

// Continuos Vesting Account

// NewContinuousVestingAccount creates a new ContinuousVestingAccount object.
func NewContinuousVestingAccount(d accountstd.Dependencies) (*ContinuousVestingAccount, error) {
	baseVestingAccount, err := NewBaseVestingAccount(d)

	continuousVestingAccount := ContinuousVestingAccount{
		BaseVestingAccount: baseVestingAccount,
		StartTime:          collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", sdk.IntValue),
	}

	return &continuousVestingAccount, err
}

type ContinuousVestingAccount struct {
	*BaseVestingAccount
	StartTime collections.Item[math.Int]
}

// --------------- Init -----------------

func (cva ContinuousVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	if msg.StartTime < 1 {
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

	return cva.BaseVestingAccount.Init(ctx, msg)
}

// --------------- execute -----------------

func (cva *ContinuousVestingAccount) ExecuteMessages(ctx context.Context, msg *account_abstractionv1.MsgExecute) (
	*account_abstractionv1.MsgExecuteResponse, error,
) {
	return cva.BaseVestingAccount.ExecuteMessages(ctx, msg, cva.GetVestingCoins)
}

// --------------- Query -----------------

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (cva ContinuousVestingAccount) GetVestedCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	endTime, err := cva.EndTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	var originalVesting sdk.Coins
	cva.IterateCoinEntries(ctx, cva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	if startTime.GTE(math.NewInt(blockTime.Unix())) {
		return vestedCoins, nil
	} else if endTime.LTE(math.NewInt(blockTime.Unix())) {
		return originalVesting, nil
	}

	// calculate the vesting scalar
	x := math.NewInt(blockTime.Unix()).Sub(startTime).Int64()
	y := endTime.Sub(startTime).Int64()
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	for _, ovc := range originalVesting {
		vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins, nil
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	var originalVesting sdk.Coins
	cva.IterateCoinEntries(ctx, cva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	vestedCoins, err := cva.GetVestedCoins(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return originalVesting.Sub(vestedCoins...), nil
}

func (cva ContinuousVestingAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	hs := cva.headerService.GetHeaderInfo(ctx)
	vestedCoins, err := cva.GetVestedCoins(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (cva ContinuousVestingAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	hs := cva.headerService.GetHeaderInfo(ctx)
	vestingCoins, err := cva.GetVestingCoins(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
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
	cva.BaseVestingAccount.RegisterExecuteHandlers(builder)
}

func (cva ContinuousVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, cva.QueryStartTime)
	accountstd.RegisterQueryHandler(builder, cva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, cva.QueryVestingCoins)
	cva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
