package upgrade

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlock will check if there is a scheduled plan and if it is ready to be executed.
// If it is ready, it will execute it if the handler is installed, and panic/abort otherwise.
// If the plan is not ready, it will ensure the handler is not registered too early (and abort otherwise).
//
// The prupose is to ensure the binary is switch EXACTLY at the desired block, and to allow
// a migration to be executed if needed upon this switch (migration defined in the new binary)
func BeginBlock(k *Keeper, ctx sdk.Context, _ abci.RequestBeginBlock) {
	plan, found := k.GetUpgradePlan(ctx)
	if !found {
		return
	}

	if plan.ShouldExecute(ctx) {
		if k.HasHandler(plan.Name) {
			// We have an upgrade handler for this upgrade name, so apply the upgrade
			ctx.Logger().Info(fmt.Sprintf("Applying upgrade \"%s\" at %s", plan.Name, plan.DueDate()))
			k.ApplyUpgrade(ctx, plan)
		} else {
			// We don't have an upgrade handler for this upgrade name, meaning this software is out of date so shutdown
			ctx.Logger().Error(fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueDate(), plan.Info))
			panic("UPGRADE REQUIRED!")
		}
	} else {
		// if we have a pending upgrade, but it is not yet time, make sure we did not
		// set the handler already
		if k.HasHandler(plan.Name) {
			ctx.Logger().Error(fmt.Sprintf("UNKNOWN UPGRADE \"%s\" - in binary but not executed on chain", plan.Name))
			panic("BINARY UPDATED BEFORE TRIGGER!")
		}
	}
}
