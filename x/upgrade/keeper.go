package upgrade

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Keeper of the upgrade module
type Keeper struct {
	storeKey        sdk.StoreKey
	cdc             *codec.Codec
	doShutdowner    func(sdk.Context, Plan)
	upgradeHandlers map[string]Handler
}

const (
	// PlanKey specifies the key under which an upgrade plan is stored in the store
	PlanKey = "plan"
)

// NewKeeper constructs an upgrade keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey:        storeKey,
		cdc:             cdc,
		upgradeHandlers: map[string]Handler{},
	}
}

// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
// must be set even if it is a no-op function.
func (keeper Keeper) SetUpgradeHandler(name string, upgradeHandler Handler) {
	keeper.upgradeHandlers[name] = upgradeHandler
}

// ScheduleUpgrade schedules an upgrade based on the specified plan
func (keeper Keeper) ScheduleUpgrade(ctx sdk.Context, plan Plan) sdk.Error {
	err := plan.ValidateBasic()
	if err != nil {
		return err
	}
	if !plan.Time.IsZero() {
		if !plan.Time.After(ctx.BlockHeader().Time) {
			return sdk.ErrUnknownRequest("Upgrade cannot be scheduled in the past")
		}
		if plan.Height != 0 {
			return sdk.ErrUnknownRequest("Only one of Time or Height should be specified")
		}
	} else {
		if plan.Height <= ctx.BlockHeight() {
			return sdk.ErrUnknownRequest("Upgrade cannot be scheduled in the past")
		}
	}
	store := ctx.KVStore(keeper.storeKey)
	if store.Has(upgradeDoneKey(plan.Name)) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("Upgrade with name %s has already been completed", plan.Name))
	}
	bz := keeper.cdc.MustMarshalBinaryBare(plan)
	store.Set([]byte(PlanKey), bz)
	return nil
}

// ClearUpgradePlan clears any schedule upgrade
func (keeper Keeper) ClearUpgradePlan(ctx sdk.Context) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete([]byte(PlanKey))
}

// ValidateBasic does basic validation of a Plan
func (plan Plan) ValidateBasic() sdk.Error {
	if len(plan.Name) == 0 {
		return sdk.ErrUnknownRequest("Name cannot be empty")

	}
	return nil
}

// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
// upgrade or false if there is none
func (keeper Keeper) GetUpgradePlan(ctx sdk.Context) (plan Plan, havePlan bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte(PlanKey))
	if bz == nil {
		return plan, false
	}
	keeper.cdc.MustUnmarshalBinaryBare(bz, &plan)
	return plan, true
}

// SetDoShutdowner sets a custom shutdown function for the upgrade module. This shutdown
// function will be called during the BeginBlock method when an upgrade is required
// instead of panic'ing which is the default behavior
func (keeper *Keeper) SetDoShutdowner(doShutdowner func(ctx sdk.Context, plan Plan)) {
	keeper.doShutdowner = doShutdowner
}

func upgradeDoneKey(name string) []byte {
	return []byte(fmt.Sprintf("done/%s", name))
}

// BeginBlocker should be called inside the BeginBlocker method of any app using the upgrade module
func (keeper *Keeper) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) {
	blockTime := ctx.BlockHeader().Time
	blockHeight := ctx.BlockHeight()

	plan, havePlan := keeper.GetUpgradePlan(ctx)
	if !havePlan {
		return
	}

	upgradeTime := plan.Time
	upgradeHeight := plan.Height
	if (!upgradeTime.IsZero() && !blockTime.Before(upgradeTime)) || upgradeHeight <= blockHeight {
		handler, ok := keeper.upgradeHandlers[plan.Name]
		if ok {
			// We have an upgrade handler for this upgrade name, so apply the upgrade
			ctx.Logger().Info(fmt.Sprintf("Applying upgrade \"%s\" at height %d", plan.Name, blockHeight))
			handler(ctx, plan)
			keeper.ClearUpgradePlan(ctx)
			// Mark this upgrade name as being done so the name can't be reused accidentally
			store := ctx.KVStore(keeper.storeKey)
			store.Set(upgradeDoneKey(plan.Name), []byte("1"))
		} else {
			// We don't have an upgrade handler for this upgrade name, meaning this software is out of date so shutdown
			ctx.Logger().Error(fmt.Sprintf("UPGRADE \"%s\" NEEDED at height %d: %s", plan.Name, blockHeight, plan.Info))
			doShutdowner := keeper.doShutdowner
			if doShutdowner != nil {
				doShutdowner(ctx, plan)
			} else {
				panic("UPGRADE REQUIRED!")
			}
		}
	}
}
