package lockup

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*ContinuousLockingAccount)(nil)
)

// NewContinuousLockingAccount creates a new ContinuousLockingAccount object.
func NewContinuousLockingAccount(clawbackEnable bool) accountstd.AccountCreatorFunc {
	return func(d accountstd.Dependencies) (string, accountstd.Interface, error) {
		if clawbackEnable {
			baseClawback := newBaseClawback(d)

			return types.CONTINUOUS_LOCKING_ACCOUNT + types.CLAWBACK_ENABLE_PREFIX, ContinuousLockingAccount{
				BaseAccount: baseClawback,
				StartTime:   collections.NewItem(d.SchemaBuilder, types.StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
			}, nil
		}

		baseLockup := newBaseLockup(d)
		return types.CONTINUOUS_LOCKING_ACCOUNT, ContinuousLockingAccount{
			BaseAccount: baseLockup,
			StartTime:   collections.NewItem(d.SchemaBuilder, types.StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
		}, nil
	}
}

type ContinuousLockingAccount struct {
	types.BaseAccount
	StartTime collections.Item[time.Time]
}

func (cva ContinuousLockingAccount) Init(ctx context.Context, msg *types.MsgInitLockupAccount) (*types.MsgInitLockupAccountResponse, error) {
	if msg.EndTime.IsZero() {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid end time %s", msg.EndTime.String())
	}

	if !msg.EndTime.After(msg.StartTime) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid start and end time (must be start before end)")
	}

	hs := cva.GetHeaderService().HeaderInfo(ctx)

	start := msg.StartTime
	if msg.StartTime.IsZero() {
		start = hs.Time
	}

	err := cva.StartTime.Set(ctx, start)
	if err != nil {
		return nil, err
	}

	return cva.BaseAccount.Init(ctx, msg, nil)
}

func (cva *ContinuousLockingAccount) Delegate(ctx context.Context, msg *types.MsgDelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := cva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not delegate enable")
	}
	return baseLockup.Delegate(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

func (cva *ContinuousLockingAccount) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (
	*types.MsgExecuteMessagesResponse, error,
) {
	baseLockup, ok := cva.BaseAccount.(*BaseLockup)
	if !ok {
		return nil, fmt.Errorf("clawback account type is not undelegate enable")
	}
	return baseLockup.Undelegate(ctx, msg)
}

func (cva *ContinuousLockingAccount) SendCoins(ctx context.Context, msg *types.MsgSend) (
	*types.MsgExecuteMessagesResponse, error,
) {
	return cva.BaseAccount.SendCoins(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

func (cva *ContinuousLockingAccount) WithdrawUnlockedCoins(ctx context.Context, msg *types.MsgWithdraw) (
	*types.MsgWithdrawResponse, error,
) {
	return cva.BaseAccount.WithdrawUnlockedCoins(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

func (cva *ContinuousLockingAccount) ClawbackFunds(ctx context.Context, msg *types.MsgClawback) (
	*types.MsgClawbackResponse, error,
) {
	baseClawback, ok := cva.BaseAccount.(*BaseClawback)
	if !ok {
		return nil, fmt.Errorf("clawback is not enable for this account type")
	}
	return baseClawback.ClawbackFunds(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

// GetLockCoinsInfo returns the total number of unlocked and locked coins.
func (cva ContinuousLockingAccount) GetLockCoinsInfo(ctx context.Context, blockTime time.Time) (unlockedCoins, lockedCoins sdk.Coins, err error) {
	unlockedCoins = sdk.Coins{}
	lockedCoins = sdk.Coins{}

	var originalVesting sdk.Coins
	err = IterateCoinEntries(ctx, cva.GetOriginalFunds(), func(key string, value math.Int) (stop bool, err error) {
		originalVesting = append(originalVesting, sdk.NewCoin(key, value))
		unlockedCoin, lockedCoin, err := cva.GetLockCoinInfoWithDenom(ctx, blockTime, key)
		if err != nil {
			return true, err
		}
		unlockedCoins = append(unlockedCoins, *unlockedCoin)
		lockedCoins = append(lockedCoins, *lockedCoin)
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return unlockedCoins, lockedCoins, nil
}

// GetLockCoinInfoWithDenom returns the number of locked coin for a specific denom. If no coins are locked,
// nil is returned.
func (cva ContinuousLockingAccount) GetLockCoinInfoWithDenom(ctx context.Context, blockTime time.Time, denom string) (unlockedCoin, lockedCoin *sdk.Coin, err error) {
	// We must handle the case where the start time for a lockup account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	endTime, err := cva.GetEndTime().Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	originalLockingAmt, err := cva.GetOriginalFunds().Get(ctx, denom)
	if err != nil {
		return nil, nil, err
	}

	originalLocking := sdk.NewCoin(denom, originalLockingAmt)
	if startTime.After(blockTime) {
		return &sdk.Coin{}, &originalLocking, nil
	} else if endTime.Before(blockTime) {
		return &originalLocking, &sdk.Coin{}, nil
	}

	// calculate the locking scalar
	x := blockTime.Unix() - startTime.Unix()
	y := endTime.Unix() - startTime.Unix()
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	unlockedAmt := math.LegacyNewDecFromInt(originalLocking.Amount).Mul(s).RoundInt()
	unlocked := sdk.NewCoin(originalLocking.Denom, unlockedAmt)

	locked := originalLocking.Sub(unlocked)

	return &unlocked, &locked, nil
}

// GetLockedCoins returns the total number of locked coins.
func (cva ContinuousLockingAccount) GetLockedCoins(ctx context.Context, blockTime time.Time) (sdk.Coins, error) {
	_, lockedCoins, err := cva.GetLockCoinsInfo(ctx, blockTime)
	if err != nil {
		return nil, err
	}
	return lockedCoins, nil
}

// GetLockedCoinsWithDenoms returns the number of locked coin for a specific denom.
func (cva ContinuousLockingAccount) GetLockedCoinsWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	lockedCoins := sdk.Coins{}
	for _, denom := range denoms {
		_, lockedCoin, err := cva.GetLockCoinInfoWithDenom(ctx, blockTime, denom)
		if err != nil {
			return nil, err
		}
		lockedCoins = append(lockedCoins, *lockedCoin)
	}

	return lockedCoins, nil
}

func (cva ContinuousLockingAccount) QueryLockupAccountInfo(ctx context.Context, req *types.QueryLockupAccountInfoRequest) (
	*types.QueryLockupAccountInfoResponse, error,
) {
	resp, err := cva.BaseAccount.QueryAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	hs := cva.GetHeaderService().HeaderInfo(ctx)
	unlockedCoins, lockedCoins, err := cva.GetLockCoinsInfo(ctx, hs.Time)
	if err != nil {
		return nil, err
	}
	resp.StartTime = &startTime
	resp.LockedCoins = lockedCoins
	resp.UnlockedCoins = unlockedCoins
	return resp, nil
}

// Implement smart account interface
func (cva ContinuousLockingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, cva.Init)
}

func (cva ContinuousLockingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, cva.Delegate)
	accountstd.RegisterExecuteHandler(builder, cva.SendCoins)
	accountstd.RegisterExecuteHandler(builder, cva.WithdrawUnlockedCoins)
	accountstd.RegisterExecuteHandler(builder, cva.ClawbackFunds)

	baseLockup, ok := cva.BaseAccount.(*BaseLockup)
	if ok {
		baseLockup.RegisterExecuteHandlers(builder)
	}
}

func (cva ContinuousLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, cva.QueryAccountBaseInfo)
}
