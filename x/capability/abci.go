package capability

import (
	"time"

	"github.com/Stride-Labs/cosmos-sdk/telemetry"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/Stride-Labs/cosmos-sdk/x/capability/keeper"
	"github.com/Stride-Labs/cosmos-sdk/x/capability/types"
)

// BeginBlocker will call InitMemStore to initialize the memory stores in the case
// that this is the first time the node is executing a block since restarting (wiping memory).
// In this case, the BeginBlocker method will reinitialize the memory stores locally, so that subsequent
// capability transactions will pass.
// Otherwise BeginBlocker performs a no-op.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.InitMemStore(ctx)
}
