package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// expected staking keeper
type StakingKeeper interface {
	IterateValidators(ctx sdk.Context,
		fn func(index int64, validator sdk.Validator) (stop bool))
}

// expected bank keeper
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(auth.Account) (stop bool))
}
