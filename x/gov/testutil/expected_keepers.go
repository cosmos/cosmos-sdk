// This file only used to generate mocks

package testutil

import (
	context "context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AccountKeeper extends gov's actual expected AccountKeeper with additional
// methods used in tests.
type AccountKeeper interface {
	types.AccountKeeper

	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
}

// BankKeeper extends gov's actual expected BankKeeper with additional
// methods used in tests.
type BankKeeper interface {
	bankkeeper.Keeper
}

// PoolKeeper extends the gov's actual expected PoolKeeper.
type PoolKeeper interface {
	types.PoolKeeper

	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// StakingKeeper extends gov's actual expected StakingKeeper with additional
// methods used in tests.
type StakingKeeper interface {
	types.StakingKeeper

	BondDenom(ctx context.Context) (string, error)
	TokensFromConsensusPower(ctx context.Context, power int64) math.Int
}
