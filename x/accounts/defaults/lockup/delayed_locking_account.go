package lockup

import (
	"context"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	lockuptypes "cosmossdk.io/x/accounts/defaults/lockup/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*DelayedLockingAccount)(nil)
)

// NewDelayedLockingAccount creates a new DelayedLockingAccount object.
func NewDelayedLockingAccount(d accountstd.Dependencies) (*DelayedLockingAccount, error) {
	baseLockup := newBaseLockup(d)
	return &DelayedLockingAccount{
		baseLockup,
	}, nil
}

type DelayedLockingAccount struct {
	*BaseLockup
}

func (dva DelayedLockingAccount) Init(ctx context.Context, msg *lockuptypes.MsgInitLockupAccount) (*lockuptypes.MsgInitLockupAccountResponse, error) {
	if msg.EndTime.IsZero() {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid end time %s", msg.EndTime.String())
	}

	return dva.BaseLockup.Init(ctx, msg)
}

func (dva *DelayedLockingAccount) Delegate(ctx context.Context, msg *lockuptypes.MsgDelegate) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return dva.BaseLockup.Delegate(ctx, msg, dva.GetLockedCoinsWithDenoms)
}

func (dva *DelayedLockingAccount) SendCoins(ctx context.Context, msg *lockuptypes.MsgSend) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return dva.BaseLockup.SendCoins(ctx, msg, dva.GetLockedCoinsWithDenoms)
}

// GetLockCoinsInfo returns the total number of unlocked and locked coins.
func (dva DelayedLockingAccount) GetLockCoinsInfo(ctx context.Context, blockTime time.Time) (sdk.Coins, sdk.Coins, error) {
	endTime, err := dva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLocking := sdk.Coins{}
	err = dva.IterateCoinEntries(ctx, dva.OriginalLocking, func(key string, value math.Int) (stop bool, err error) {
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
	endTime, err := dva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	originalLockingAmt, err := dva.OriginalLocking.Get(ctx, denom)
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

func (dva DelayedLockingAccount) QueryVestingAccountInfo(ctx context.Context, req *lockuptypes.QueryLockupAccountInfoRequest) (
	*lockuptypes.QueryLockupAccountInfoResponse, error,
) {
	resp, err := dva.BaseLockup.QueryLockupAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	hs := dva.headerService.HeaderInfo(ctx)
	unlockedCoins, lockedCoins, err := dva.GetLockCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.LockedCoins = lockedCoins
	resp.UnlockedCoins = unlockedCoins
	return resp, nil
}

func (dva DelayedLockingAccount) QuerySpendableTokens(ctx context.Context, req *lockuptypes.QuerySpendableAmountRequest) (
	*lockuptypes.QuerySpendableAmountResponse, error,
) {
	hs := dva.headerService.HeaderInfo(ctx)
	_, lockedCoins, err := dva.GetLockCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}

	resp, err := dva.BaseLockup.QuerySpendableTokens(ctx, lockedCoins)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Implement smart account interface
func (dva DelayedLockingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, dva.Init)
}

func (dva DelayedLockingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, dva.Delegate)
	accountstd.RegisterExecuteHandler(builder, dva.SendCoins)
	dva.BaseLockup.RegisterExecuteHandlers(builder)
}

func (dva DelayedLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, dva.QueryVestingAccountInfo)
	accountstd.RegisterQueryHandler(builder, dva.QuerySpendableTokens)
	dva.BaseLockup.RegisterQueryHandlers(builder)
}
