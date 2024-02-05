package vesting

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
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
		StartTime:   collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return &continuousVestingAccount, nil
}

type ContinuousVestingAccount struct {
	*BaseVesting
	StartTime collections.Item[time.Time]
}

func (cva ContinuousVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	if msg.EndTime.IsZero() {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid end time %s", msg.EndTime.String())
	}

	if msg.EndTime.Before(msg.StartTime) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid start and end time (must be start before end)")
	}

	hs := cva.headerService.GetHeaderInfo(ctx)

	start := msg.StartTime
	if msg.StartTime.IsZero() {
		start = hs.Time
	}

	err := cva.StartTime.Set(ctx, start)
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
	err = cva.IterateCoinEntries(ctx, cva.OriginalVesting, func(key string, value math.Int) (stop bool, err error) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if startTime.After(blockTime) {
		return vestedCoins, originalVesting, nil
	} else if endTime.Before(blockTime) {
		return originalVesting, vestingCoins, nil
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - startTime.Unix()
	y := endTime.Unix() - startTime.Unix()
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

func (cva ContinuousVestingAccount) QueryVestingAccountInfo(ctx context.Context, req *vestingtypes.QueryVestingAccountInfoRequest) (
	*vestingtypes.QueryVestingAccountInfoResponse, error,
) {
	resp, err := cva.BaseVesting.QueryVestingAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	hs := cva.headerService.GetHeaderInfo(ctx)
	vestedCoins, vestingCoins, err := cva.GetVestCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.VestedVesting = sdk.Coins{}
	resp.StartTime = &startTime
	resp.VestingCoins = vestingCoins
	resp.VestedVesting = vestedCoins
	return resp, nil
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
	accountstd.RegisterQueryHandler(builder, cva.QueryVestingAccountInfo)
}
