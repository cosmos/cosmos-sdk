package lockup

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*DelayedLockingAccount)(nil)
)

// NewDelayedLockingAccount creates a new DelayedLockingAccount object.
func NewDelayedLockingAccount(clawbackEnable bool) accountstd.AccountCreatorFunc {
	return func(d accountstd.Dependencies) (string, accountstd.Interface, error) {
		if clawbackEnable {
			baseClawback := newBaseClawback(d)

			return types.DELAYED_LOCKING_ACCOUNT + types.CLAWBACK_ENABLE_PREFIX, DelayedLockingAccount{
				BaseAccount: baseClawback,
			}, nil
		}

		baseLockup := newBaseLockup(d)
		return types.DELAYED_LOCKING_ACCOUNT, DelayedLockingAccount{
			BaseAccount: baseLockup,
		}, nil
	}
}

type DelayedLockingAccount struct {
	types.BaseAccount
}

func (dva DelayedLockingAccount) Init(ctx context.Context, msg *types.MsgInitLockupAccount) (*types.MsgInitLockupAccountResponse, error) {
	if msg.EndTime.IsZero() {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid end time %s", msg.EndTime.String())
	}

	return dva.BaseAccount.Init(ctx, msg, nil)
}

func (dva *DelayedLockingAccount) Delegate(ctx context.Context, msg *types.MsgDelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := dva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not delegate enable")
	}
	return baseLockup.Delegate(ctx, msg, dva.GetLockedCoinsWithDenoms)
}

func (dva *DelayedLockingAccount) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := dva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not undelegate enable")
	}
	return baseLockup.Undelegate(ctx, msg)
}

func (dva *DelayedLockingAccount) SendCoins(ctx context.Context, msg *types.MsgSend) (
	*types.MsgExecuteMessagesResponse, error,
) {
	return dva.BaseAccount.SendCoins(ctx, msg, dva.GetLockedCoinsWithDenoms)
}

func (dva *DelayedLockingAccount) WithdrawUnlockedCoins(ctx context.Context, msg *types.MsgWithdraw) (
	*types.MsgWithdrawResponse, error,
) {
	return dva.BaseAccount.WithdrawUnlockedCoins(ctx, msg, dva.GetLockedCoinsWithDenoms)
}

func (dva *DelayedLockingAccount) ClawbackFunds(ctx context.Context, msg *types.MsgClawback) (
	*types.MsgClawbackResponse, error,
) {
	baseClawback, ok := dva.BaseAccount.(*BaseClawback)
	if !ok {
		return nil, fmt.Errorf("clawback is not enable for this account type")
	}
	return baseClawback.ClawbackFunds(ctx, msg, dva.GetLockedCoinsWithDenoms)
}

// GetLockCoinsInfo returns the total number of unlocked and locked coins.
func (dva DelayedLockingAccount) GetLockCoinsInfo(ctx context.Context, blockTime time.Time) (sdk.Coins, sdk.Coins, error) {
	endTime, err := dva.GetEndTime().Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLocking := sdk.Coins{}
	err = IterateCoinEntries(ctx, dva.GetOriginalFunds(), func(key string, value math.Int) (stop bool, err error) {
		originalLocking = append(originalLocking, sdk.NewCoin(key, value))
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if blockTime.After(endTime) {
		return originalLocking, sdk.Coins{}, nil
	}

	return sdk.Coins{}, originalLocking, nil
}

// GetLockedCoins returns the total number of locked coins. If no coins are
// locked, nil is returned.
func (dva DelayedLockingAccount) GetLockedCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	_, lockedCoins, err := dva.GetLockCoinsInfo(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return lockedCoins, nil
}

// GetLockCoinInfoWithDenom returns the number of unlocked and locked coin for a specific denom.
func (dva DelayedLockingAccount) GetLockCoinInfoWithDenom(ctx context.Context, blockTime time.Time, denom string) (*sdk.Coin, *sdk.Coin, error) {
	endTime, err := dva.GetEndTime().Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLockingAmt, err := dva.GetOriginalFunds().Get(ctx, denom)
	if err != nil {
		return nil, nil, err
	}
	originalLockingCoin := sdk.NewCoin(denom, originalLockingAmt)

	if blockTime.After(endTime) {
		return &originalLockingCoin, &sdk.Coin{}, nil
	}

	return &sdk.Coin{}, &originalLockingCoin, nil
}

// GetLockedCoinsWithDenoms returns the number of locked coin for a specific denom.
func (dva DelayedLockingAccount) GetLockedCoinsWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	vestingCoins := sdk.Coins{}
	for _, denom := range denoms {
		_, vestingCoin, err := dva.GetLockCoinInfoWithDenom(ctx, blockTime, denom)
		if err != nil {
			return nil, err
		}
		vestingCoins = append(vestingCoins, *vestingCoin)
	}
	return vestingCoins, nil
}

func (dva DelayedLockingAccount) QueryVestingAccountInfo(ctx context.Context, req *types.QueryLockupAccountInfoRequest) (
	*types.QueryLockupAccountInfoResponse, error,
) {
	resp, err := dva.BaseAccount.QueryAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	hs := dva.GetHeaderService().HeaderInfo(ctx)
	unlockedCoins, lockedCoins, err := dva.GetLockCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.LockedCoins = lockedCoins
	resp.UnlockedCoins = unlockedCoins
	return resp, nil
}

// Implement smart account interface
func (dva DelayedLockingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, dva.Init)
}

func (dva DelayedLockingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, dva.Delegate)
	accountstd.RegisterExecuteHandler(builder, dva.SendCoins)
	accountstd.RegisterExecuteHandler(builder, dva.WithdrawUnlockedCoins)
	accountstd.RegisterExecuteHandler(builder, dva.ClawbackFunds)
	dva.BaseLockup.RegisterExecuteHandlers(builder)
}

func (dva DelayedLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, dva.QueryVestingAccountInfo)
}
