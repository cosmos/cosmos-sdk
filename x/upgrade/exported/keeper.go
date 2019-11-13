package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/internal/types"
)

// Keeper of the upgrade module
type Keeper interface {
	// ScheduleUpgrade schedules an upgrade based on the specified plan
	ScheduleUpgrade(ctx sdk.Context, plan types.Plan) sdk.Error

	// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
	// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
	// must be set even if it is a no-op function.
	SetUpgradeHandler(name string, upgradeHandler types.UpgradeHandler)

	// ClearUpgradePlan clears any schedule upgrade
	ClearUpgradePlan(ctx sdk.Context)

	// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
	// upgrade or false if there is none
	GetUpgradePlan(ctx sdk.Context) (plan types.Plan, havePlan bool)

	// HasHandler returns true iff there is a handler registered for this name
	HasHandler(name string) bool

	// ApplyUpgrade will execute the handler associated with the Plan and mark the plan as done.
	ApplyUpgrade(ctx sdk.Context, plan types.Plan)
}
