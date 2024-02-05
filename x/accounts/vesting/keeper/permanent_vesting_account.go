package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*PermanentLockedAccount)(nil)
)

// NewPermanentLockedAccount creates a new PermanentLockedAccount object.
func NewPermanentLockedAccount(d accountstd.Dependencies) (*PermanentLockedAccount, error) {
	baseVesting := NewBaseVesting(d)

	return &PermanentLockedAccount{baseVesting}, nil
}

type PermanentLockedAccount struct {
	*BaseVesting
}

func (plva PermanentLockedAccount) Init(ctx context.Context, msg *vestingtypes.MsgInitVestingAccount) (*vestingtypes.MsgInitVestingAccountResponse, error) {
	resp, err := plva.BaseVesting.Init(ctx, msg)
	if err != nil {
		return nil, err
	}
	err = plva.EndTime.Set(ctx, time.Time{})
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (plva *PermanentLockedAccount) ExecuteMessages(ctx context.Context, msg *account_abstractionv1.MsgExecute) (
	*account_abstractionv1.MsgExecuteResponse, error,
) {
	return plva.BaseVesting.ExecuteMessages(ctx, msg, func(_ context.Context, _ time.Time) (sdk.Coins, error) {
		originalVesting := sdk.Coins{}
		err := plva.IterateCoinEntries(ctx, plva.OriginalVesting, func(key string, value math.Int) (stop bool, err error) {
			originalVesting = append(originalVesting, sdk.NewCoin(key, value))
			return false, nil
		})
		if err != nil {
			return nil, err
		}
		return originalVesting, nil
	})
}

func (plva PermanentLockedAccount) QueryVestingAccountInfo(ctx context.Context, req *vestingtypes.QueryVestingAccountInfoRequest) (
	*vestingtypes.QueryVestingAccountInfoResponse, error,
) {
	resp, err := plva.BaseVesting.QueryVestingAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	originalVesting := sdk.Coins{}
	plva.IterateCoinEntries(ctx, plva.OriginalVesting, func(key string, value math.Int) (stop bool, err error) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		return false, nil
	})
	resp.VestingCoins = originalVesting
	resp.VestedVesting = sdk.Coins{}
	return resp, nil
}

// Implement smart account interface
func (plva PermanentLockedAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, plva.Init)
}

func (plva PermanentLockedAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, plva.ExecuteMessages)
	plva.BaseVesting.RegisterExecuteHandlers(builder)
}

func (plva PermanentLockedAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, plva.QueryVestingAccountInfo)
}
