package vesting

import (
	"context"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*DelayedVestingAccount)(nil)
)

// NewDelayedVestingAccount creates a new DelayedVestingAccount object.
func NewDelayedVestingAccount(d accountstd.Dependencies) (*DelayedVestingAccount, error) {
	baseVesting := NewBaseVesting(d)
	return &DelayedVestingAccount{
		baseVesting,
	}, nil
}

type DelayedVestingAccount struct {
	*BaseVesting
}

func (dva DelayedVestingAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	if msg.EndTime.IsZero() {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid end time %s", msg.EndTime.String())
	}

	return dva.BaseVesting.Init(ctx, msg)
}

func (dva *DelayedVestingAccount) ExecuteMessages(ctx context.Context, msg *account_abstractionv1.MsgExecute) (
	*account_abstractionv1.MsgExecuteResponse, error,
) {
	return dva.BaseVesting.ExecuteMessages(ctx, msg, dva.GetVestingCoins)
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (dva DelayedVestingAccount) GetVestCoinsInfo(ctx context.Context, blockTime time.Time) (sdk.Coins, sdk.Coins, error) {
	endTime, err := dva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalVesting := sdk.Coins{}
	dva.IterateCoinEntries(ctx, dva.OriginalVesting, func(key string, value math.Int) (stop bool) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false
	})
	if blockTime.After(endTime) {
		return originalVesting, sdk.Coins{}, nil
	}

	return sdk.Coins{}, originalVesting, nil
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (dva DelayedVestingAccount) GetVestingCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	_, vestingCoins, err := dva.GetVestCoinsInfo(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return vestingCoins, nil
}

func (dva DelayedVestingAccount) QueryVestingAccountInfo(ctx context.Context, req *vestingtypes.QueryVestingAccountInfoRequest) (
	*vestingtypes.QueryVestingAccountInfoResponse, error,
) {
	resp, err := dva.BaseVesting.QueryVestingAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	hs := dva.headerService.GetHeaderInfo(ctx)
	vestedCoins, vestingCoins, err := dva.GetVestCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.VestedVesting = sdk.Coins{}
	resp.VestingCoins = vestingCoins
	resp.VestedVesting = vestedCoins
	return resp, nil
}

// Implement smart account interface
func (dva DelayedVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, dva.Init)
}

func (dva DelayedVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, dva.ExecuteMessages)
	dva.BaseVesting.RegisterExecuteHandlers(builder)
}

func (dva DelayedVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, dva.QueryVestingAccountInfo)
}
