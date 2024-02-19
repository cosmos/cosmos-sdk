package vesting

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
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

func (cva *ContinuousVestingAccount) ExecuteMessages(ctx context.Context, msg *vestingtypes.MsgExecuteMessages) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	return cva.BaseVesting.ExecuteMessages(ctx, msg, cva.GetVestingCoinWithDenoms)
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
		vestedCoin, vestingCoin, err := cva.GetVestCoinInfoWithDenom(ctx, blockTime, key)
		if err != nil {
			return true, err
		}
		vestedCoins = append(vestedCoins, *vestedCoin)
		vestingCoins = append(vestingCoins, *vestingCoin)
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

	return vestedCoins, vestingCoins, nil
}

// GetVestCoinsInfoWithDenom returns the number of vested coin for a specific denom. If no coins are vested,
// nil is returned.
func (cva ContinuousVestingAccount) GetVestCoinInfoWithDenom(ctx context.Context, blockTime time.Time, denom string) (vestedCoin, vestingCoin *sdk.Coin, err error) {
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

	originalVestingAmt, err := cva.OriginalVesting.Get(ctx, denom)
	if err != nil {
		return nil, nil, err
	}

	originalVesting := sdk.NewCoin(denom, originalVestingAmt)
	if startTime.After(blockTime) {
		return nil, &originalVesting, nil
	} else if endTime.Before(blockTime) {
		return &originalVesting, nil, nil
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - startTime.Unix()
	y := endTime.Unix() - startTime.Unix()
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	vestedAmt := math.LegacyNewDecFromInt(originalVesting.Amount).Mul(s).RoundInt()
	vested := sdk.NewCoin(originalVesting.Denom, vestedAmt)

	vesting := originalVesting.Sub(vested)

	return &vested, &vesting, nil
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

// GetVestingCoinsWithDenom returns the number of vesting coin for a specific denom. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoinWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	vestingCoins := sdk.Coins{}
	for _, denom := range denoms {
		_, vestingCoin, err := cva.GetVestCoinInfoWithDenom(ctx, blockTime, denom)
		if err != nil {
			return nil, err
		}
		vestingCoins = append(vestingCoins, *vestingCoin)
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
}

func (cva ContinuousVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, cva.QueryVestingAccountInfo)
}
