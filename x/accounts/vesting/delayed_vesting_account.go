package vesting

import (
	"context"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*DelayedVestingAccount)(nil)
)

// Delayed Vesting Account

// NewDelayedVestingAccount creates a new DelayedVestingAccount object.
func NewDelayedVestingAccount(d accountstd.Dependencies) (*DelayedVestingAccount, error) {
	baseVestingAccount, err := NewBaseVestingAccount(d)
	return &DelayedVestingAccount{
		baseVestingAccount,
	}, err
}

type DelayedVestingAccount struct {
	*BaseVestingAccount
}

// --------------- Init -----------------

func (dva DelayedVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	if msg.EndTime <= 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid end time")
	}

	return dva.BaseVestingAccount.Init(ctx, msg)
}

// --------------- execute -----------------

func (dva *DelayedVestingAccount) ExecuteMessages(ctx context.Context, msg *vestingtypes.MsgExecuteMessages) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	return dva.BaseVestingAccount.ExecuteMessages(ctx, msg, dva.GetVestingCoins)
}

// --------------- Query -----------------

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (dva DelayedVestingAccount) GetVestedCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	endTime, err := dva.EndTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	var originalVesting sdk.Coins
	dva.IterateCoinEntries(ctx, dva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	if math.NewInt(blockTime.Unix()).GTE(endTime) {
		return originalVesting, nil
	}

	return nil, nil
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (dva DelayedVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	var originalVesting sdk.Coins
	dva.IterateCoinEntries(ctx, dva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	vestedCoins, err := dva.GetVestedCoins(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return originalVesting.Sub(vestedCoins...), nil
}

func (dva DelayedVestingAccount) QueryVestedCoins(ctx context.Context, msg *vestingtypes.QueryVestedCoinsRequest) (
	*vestingtypes.QueryVestedCoinsResponse, error,
) {
	hs := dva.headerService.GetHeaderInfo(ctx)
	vestedCoins, err := dva.GetVestedCoins(ctx, hs.Time)
	if err != nil {
		return nil, err
	}

	return &vestingtypes.QueryVestedCoinsResponse{
		VestedVesting: vestedCoins,
	}, nil
}

func (dva DelayedVestingAccount) QueryVestingCoins(ctx context.Context, msg *vestingtypes.QueryVestingCoinsRequest) (
	*vestingtypes.QueryVestingCoinsResponse, error,
) {
	hs := dva.headerService.GetHeaderInfo(ctx)
	vestingCoins, err := dva.GetVestingCoins(ctx, hs.Time)
	if err != nil {
		return nil, err
	}

	return &vestingtypes.QueryVestingCoinsResponse{
		VestingCoins: vestingCoins,
	}, nil
}

// Implement smart account interface
func (dva DelayedVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, dva.Init)
}

func (dva DelayedVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, dva.ExecuteMessages)
}

func (dva DelayedVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, dva.QueryVestedCoins)
	accountstd.RegisterQueryHandler(builder, dva.QueryVestingCoins)
	dva.BaseVestingAccount.RegisterQueryHandlers(builder)
}
