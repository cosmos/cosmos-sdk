package slashing

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected staking keeper
type StakingKeeper interface {
	IterateValidators(ctx sdk.Context,
		fn func(index int64, validator sdk.Validator) (stop bool))
}
