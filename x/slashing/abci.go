package slashing

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var deprecatedBitArrayPruneLimitPerBlock = 2000

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	params := k.GetParams(ctx)
	for _, voteInfo := range req.LastCommitInfo.GetVotes() {
		k.HandleValidatorSignatureWithParams(ctx, params, voteInfo.Validator.Address, voteInfo.Validator.Power, voteInfo.SignedLastBlock)
	}

	// If there are still entries for the deprecated MissedBlockBitArray, delete them up until we hit the per block limit
	k.DeleteDeprecatedValidatorMissedBlockBitArray(ctx, deprecatedBitArrayPruneLimitPerBlock)
}
