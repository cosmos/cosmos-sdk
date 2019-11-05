package upgrade

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock will check if there is a scheduled plan and if it is ready to be executed.
// If skip upgrade flag is set, it will skip and clear the upgrade plan
// If it is ready, it will execute it if the handler is installed, and panic/abort otherwise.
// If the plan is not ready, it will ensure the handler is not registered too early (and abort otherwise).
//
// The purpose is to ensure the binary is switched EXACTLY at the desired block, and to allow
// a migration to be executed if needed upon this switch (migration defined in the new binary)
func BeginBlocker(k Keeper, ctx sdk.Context, _ abci.RequestBeginBlock, skipUpgrade bool) {

	plan, found := k.GetUpgradePlan(ctx)
	if !found {
		return
	}

	if plan.ShouldExecute(ctx) {
		//To make sure clear upgrade is executed at the same block
		if skipUpgrade {
			// If skip upgrade has been set, we clear the upgrade plan
			skipUpgradeMsg := fmt.Sprintf("UPGRADE \"%s\" SKIPPED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
			ctx.Logger().Info(skipUpgradeMsg)
			k.ClearUpgradePlan(ctx)
			return
		}

		if !k.HasHandler(plan.Name) {
			upgradeMsg := fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
			// We don't have an upgrade handler for this upgrade name, meaning this software is out of date so shutdown
			ctx.Logger().Error(upgradeMsg)
			panic(upgradeMsg)
		}
		// We have an upgrade handler for this upgrade name, so apply the upgrade
		ctx.Logger().Info(fmt.Sprintf("applying upgrade \"%s\" at %s", plan.Name, plan.DueAt()))
		ctx = ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())
		k.ApplyUpgrade(ctx, plan)
		return
	}

	// if we have a pending upgrade, but it is not yet time, make sure we did not
	// set the handler already
	if k.HasHandler(plan.Name) {
		downgradeMsg := fmt.Sprintf("BINARY UPDATED BEFORE TRIGGER! UPGRADE \"%s\" - in binary but not executed on chain", plan.Name)
		ctx.Logger().Error(downgradeMsg)
		panic(downgradeMsg)
	}
}
