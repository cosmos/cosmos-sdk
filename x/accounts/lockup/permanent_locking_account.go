package lockup

import (
	"context"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	lockuptypes "cosmossdk.io/x/accounts/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*PermanentLockingAccount)(nil)
)

// NewPermanentLockingAccount creates a new PermanentLockingAccount object.
func NewPermanentLockingAccount(d accountstd.Dependencies) (*PermanentLockingAccount, error) {
	baseLockup := newBaseLockup(d)

	return &PermanentLockingAccount{baseLockup}, nil
}

type PermanentLockingAccount struct {
	*BaseLockup
}

func (plva PermanentLockingAccount) Init(ctx context.Context, msg *lockuptypes.MsgInitLockupAccount) (*lockuptypes.MsgInitLockupAccountResponse, error) {
	resp, err := plva.BaseLockup.Init(ctx, msg)
	if err != nil {
		return nil, err
	}
	err = plva.EndTime.Set(ctx, time.Time{})
	if err != nil {
		return nil, err
	}

	return resp, err
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (plva PermanentLockingAccount) GetVestingCoinWithDenoms(ctx context.Context, blockTime time.Time, denoms ...string) (sdk.Coins, error) {
	vestingCoins := sdk.Coins{}
	for _, denom := range denoms {
		originalVestingAmt, err := plva.OriginalLocking.Get(ctx, denom)
		if err != nil {
			return nil, err
		}
		vestingCoins = append(vestingCoins, sdk.NewCoin(denom, originalVestingAmt))
	}
	return vestingCoins, nil
}

func (plva *PermanentLockingAccount) ExecuteMessages(ctx context.Context, msg *lockuptypes.MsgExecuteMessages) (
	*lockuptypes.MsgExecuteMessagesResponse, error,
) {
	return plva.BaseLockup.ExecuteMessages(ctx, msg, plva.GetVestingCoinWithDenoms)
}

func (plva PermanentLockingAccount) QueryLockupAccountInfo(ctx context.Context, req *lockuptypes.QueryLockupAccountInfoRequest) (
	*lockuptypes.QueryLockupAccountInfoResponse, error,
) {
	resp, err := plva.BaseLockup.QueryLockupAccountBaseInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	originalLocking := sdk.Coins{}
	err = plva.IterateCoinEntries(ctx, plva.OriginalLocking, func(key string, value math.Int) (stop bool, err error) {
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
	accountstd.RegisterExecuteHandler(builder, plva.ExecuteMessages)
}

func (plva PermanentLockingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, plva.QueryLockupAccountInfo)
}
