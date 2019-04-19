package genutil

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected staking keeper
type StakingKeeper interface {
	ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) (updates []abci.ValidatorUpdate)
}
