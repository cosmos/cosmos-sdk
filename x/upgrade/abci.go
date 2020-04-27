package upgrade

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock will check if there is a scheduled plan and if it is ready to be executed.
// If the current height is in the provided set of heights to skip, it will skip and clear the upgrade plan.
// If it is ready, it will execute it if the handler is installed, and panic/abort otherwise.
// If the plan is not ready, it will ensure the handler is not registered too early (and abort otherwise).
//
// The purpose is to ensure the binary is switched EXACTLY at the desired block, and to allow
// a migration to be executed if needed upon this switch (migration defined in the new binary)
// skipUpgradeHeightArray is a set of block heights for which the upgrade must be skipped
func BeginBlocker(k Keeper, ctx sdk.Context, _ abci.RequestBeginBlock) {
	plan, found := k.GetUpgradePlan(ctx)
	if !found {
		return
	}

	// to make sure clear upgrade is executed at the same block
	if plan.ShouldExecute(ctx) {
		// if skip upgrade has been set for current height, we clear the upgrade plan
		if k.IsSkipHeight(ctx.BlockHeight()) {
			skipUpgradeMsg := fmt.Sprintf("UPGRADE \"%s\" SKIPPED at %d: %s", plan.Name, plan.Height, plan.Info)
			k.Logger(ctx).Info(skipUpgradeMsg)

			// clear the upgrade plan at current height
			k.ClearUpgradePlan(ctx)
			return
		}

		if !k.HasHandler(plan.Name) {
			upgradeMsg := fmt.Sprintf("UPGRADE \"%s\" NEEDED at %s: %s", plan.Name, plan.DueAt(), plan.Info)
			// we don't have an upgrade handler for this upgrade name, meaning this software is out of date so shutdown
			k.Logger(ctx).Error(upgradeMsg)

			// Write the upgrade info to disk. The UpgradeStoreLoader uses this info to perform or skip
			// store migrations.
			err := k.DumpUpgradeInfoToDisk(ctx.BlockHeight(), plan.Name)
			if err != nil {
				panic(fmt.Errorf("unable to write upgrade info to filesystem: %s", err.Error()))
			}

			panic(upgradeMsg)
		}

		// we have an upgrade handler for this upgrade name, so apply the upgrade
		k.Logger(ctx).Info(fmt.Sprintf("applying upgrade \"%s\" at %s", plan.Name, plan.DueAt()))

		// TODO: Document why we use an infinite gas meter.
		ctx = ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())
		k.ApplyUpgrade(ctx, plan)

		return
	}

	// If we have a pending upgrade, but it is not yet time, make sure we did not
	// set the handler already.
	if k.HasHandler(plan.Name) {
		downgradeMsg := fmt.Sprintf("BINARY UPDATED BEFORE TRIGGER! UPGRADE \"%s\" - in binary but not executed on chain", plan.Name)
		k.Logger(ctx).Error(downgradeMsg)
		panic(downgradeMsg)
	}
}
