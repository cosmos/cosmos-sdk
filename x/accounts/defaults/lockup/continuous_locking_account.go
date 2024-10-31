package lockup

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	lockuptypes "cosmossdk.io/x/accounts/defaults/lockup/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*ContinuousLockingAccount)(nil)
)

// NewContinuousLockingAccount creates a new ContinuousLockingAccount object.
func NewContinuousLockingAccount(d accountstd.Dependencies) (*ContinuousLockingAccount, error) {
	baseLockup := newBaseLockup(d)

	ContinuousLockingAccount := ContinuousLockingAccount{
		BaseLockup: baseLockup,
		StartTime:  collections.NewItem(d.SchemaBuilder, StartTimePrefix, "start_time", collcodec.KeyToValueCodec[time.Time](sdk.TimeKey)),
	}

	return &ContinuousLockingAccount, nil
}

type ContinuousLockingAccount struct {
	*BaseLockup
	StartTime collections.Item[time.Time]
}

func (cva ContinuousLockingAccount) Init(ctx context.Context, msg *lockuptypes.MsgInitLockupAccount) (*lockuptypes.MsgInitLockupAccountResponse, error) {
	if msg.EndTime.IsZero() {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid end time %s", msg.EndTime.String())
	}

	if !msg.EndTime.After(msg.StartTime) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid start and end time (must be start before end)")
	}

	hs := cva.headerService.HeaderInfo(ctx)

	start := msg.StartTime
	if msg.StartTime.IsZero() {
		start = hs.Time
	}

	err := cva.StartTime.Set(ctx, start)
	if err != nil {
		return nil, err
	}

	return cva.BaseLockup.Init(ctx, msg)
}

func (cva *ContinuousLockingAccount) Delegate(ctx context.Context, msg *lockuptypes.MsgDelegate) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return cva.BaseLockup.Delegate(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

func (cva *ContinuousLockingAccount) SendCoins(ctx context.Context, msg *lockuptypes.MsgSend) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return cva.BaseLockup.SendCoins(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

func (cva *ContinuousLockingAccount) WithdrawUnlockedCoins(ctx context.Context, msg *lockuptypes.MsgWithdraw) (
	*lockuptypes.MsgWithdrawResponse, error,
) {
	return cva.BaseLockup.WithdrawUnlockedCoins(ctx, msg, cva.GetLockedCoinsWithDenoms)
}

// GetLockCoinsInfo returns the total number of unlocked and locked coins.
func (cva ContinuousLockingAccount) GetLockCoinsInfo(ctx context.Context, blockTime time.Time) (unlockedCoins, lockedCoins sdk.Coins, err error) {
	unlockedCoins = sdk.Coins{}
	lockedCoins = sdk.Coins{}

	var originalLocking sdk.Coins
	err = cva.IterateCoinEntries(ctx, cva.OriginalLocking, func(key string, value math.Int) (stop bool, err error) {
		originalLocking = append(originalLocking, sdk.NewCoin(key, value))
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
	endTime, err := cva.EndTime.Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	originalLockingAmt, err := cva.OriginalLocking.Get(ctx, denom)
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

func (cva ContinuousLockingAccount) QueryLockupAccountInfo(ctx context.Context, req *lockuptypes.QueryLockupAccountInfoRequest) (
	*lockuptypes.QueryLockupAccountInfoResponse, error,
) {
	resp, err := cva.BaseLockup.QueryLockupAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	startTime, err := cva.StartTime.Get(ctx)
	if err != nil {
		return nil, err
	}
	hs := cva.headerService.HeaderInfo(ctx)
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
	cva.BaseLockup.RegisterExecuteHandlers(builder)
}

func (cva ContinuousLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, cva.QueryLockupAccountInfo)
}
