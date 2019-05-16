package crisis

import (
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// check all registered invariants
func EndBlocker(ctx sdk.Context, k Keeper, logger log.Logger) {
	if k.invCheckPeriod == 0 || ctx.BlockHeight()%int64(k.invCheckPeriod) != 0 {
		// skip running the invariant check
		return
	}
	k.AssertInvariants(ctx, logger)
}
