package lockup

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*PermanentLockingAccount)(nil)
)

// NewPermanentLockingAccount creates a new PermanentLockingAccount object.
func NewPermanentLockingAccount(clawbackEnable bool) accountstd.AccountCreatorFunc {
	return func(d accountstd.Dependencies) (string, accountstd.Interface, error) {
		if clawbackEnable {
			baseClawback := newBaseClawback(d)

			return types.PERMANENT_LOCKING_ACCOUNT + types.CLAWBACK_ENABLE_PREFIX, PermanentLockingAccount{
				BaseAccount: baseClawback,
			}, nil
		}

		baseLockup := newBaseLockup(d)
		return types.PERMANENT_LOCKING_ACCOUNT, PermanentLockingAccount{
			BaseAccount: baseLockup,
		}, nil
	}
}

type PermanentLockingAccount struct {
	types.BaseAccount
}

func (plva PermanentLockingAccount) Init(ctx context.Context, msg *types.MsgInitLockupAccount) (*types.MsgInitLockupAccountResponse, error) {
	resp, err := plva.BaseAccount.Init(ctx, msg, nil)
	if err != nil {
		return nil, err
	}
	err = plva.GetEndTime().Set(ctx, time.Time{})
	if err != nil {
		return nil, err
	}

	return resp, err
}

// GetlockedCoinsWithDenoms returns the total number of locked coins. If no coins are
// locked, nil is returned.
func (plva PermanentLockingAccount) GetLockedCoinsWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	vestingCoins := sdk.Coins{}
	for _, denom := range denoms {
		originalVestingAmt, err := plva.GetOriginalFunds().Get(ctx, denom)
		if err != nil {
			return nil, err
		}
		vestingCoins = append(vestingCoins, sdk.NewCoin(denom, originalVestingAmt))
	}
	return vestingCoins, nil
}

func (plva *PermanentLockingAccount) Delegate(ctx context.Context, msg *types.MsgDelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := plva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not delegate enable")
	}
	return baseLockup.Delegate(ctx, msg, plva.GetLockedCoinsWithDenoms)
}

func (plva *PermanentLockingAccount) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := plva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not undelegate enable")
	}
	return baseLockup.Undelegate(ctx, msg)
}

func (plva *PermanentLockingAccount) SendCoins(ctx context.Context, msg *types.MsgSend) (
	*types.MsgExecuteMessagesResponse, error,
) {
	return plva.BaseAccount.SendCoins(ctx, msg, plva.GetLockedCoinsWithDenoms)
}

func (plva *PermanentLockingAccount) ClawbackFunds(ctx context.Context, msg *types.MsgClawback) (
	*types.MsgClawbackResponse, error,
) {
	baseClawback, ok := plva.BaseAccount.(*BaseClawback)
	if !ok {
		return nil, fmt.Errorf("clawback is not enable for this account type")
	}
	return baseClawback.ClawbackFunds(ctx, msg, plva.GetLockedCoinsWithDenoms)
}

func (plva PermanentLockingAccount) QueryLockupAccountInfo(ctx context.Context, req *types.QueryLockupAccountInfoRequest) (
	*types.QueryLockupAccountInfoResponse, error,
) {
	resp, err := plva.BaseAccount.QueryAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	originalLocking := sdk.Coins{}
	err = IterateCoinEntries(ctx, plva.GetOriginalFunds(), func(key string, value math.Int) (stop bool, err error) {
		originalLocking = append(originalLocking, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	resp.LockedCoins = originalLocking
	resp.UnlockedCoins = sdk.Coins{}
	return resp, nil
}

// Implement smart account interface
func (plva PermanentLockingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, plva.Init)
}

func (plva PermanentLockingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, plva.Delegate)
	accountstd.RegisterExecuteHandler(builder, plva.SendCoins)
	accountstd.RegisterExecuteHandler(builder, plva.ClawbackFunds)
	plva.BaseLockup.RegisterExecuteHandlers(builder)
}

func (plva PermanentLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, plva.QueryLockupAccountInfo)
}
