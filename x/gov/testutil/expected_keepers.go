// This file only used to generate mocks

package testutil

import (
	math "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AccountKeeper is the same interface as gov's actual expected AccountKeeper.
type AccountKeeper interface {
	types.AccountKeeper
}

// BankKeeper extends gov's actual expected BankKeeper with additional
// methods used in tests.
type BankKeeper interface {
	types.BankKeeper

	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper extends gov's actual expected StakingKeeper with additional
// methods used in tests.
type StakingKeeper interface {
	types.StakingKeeper

	TokensFromConsensusPower(ctx sdk.Context, power int64) math.Int
}
