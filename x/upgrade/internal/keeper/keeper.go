package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/exported"
	"github.com/cosmos/cosmos-sdk/x/upgrade/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Keeper struct {
	storeKey        sdk.StoreKey
	cdc             *codec.Codec
	upgradeHandlers map[string]types.UpgradeHandler
	haveCache       bool
}

var _ exported.Keeper = (*Keeper)(nil)

// NewKeeper constructs an upgrade Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) *Keeper {
	return &Keeper{
		storeKey:        storeKey,
		cdc:             cdc,
		upgradeHandlers: map[string]types.UpgradeHandler{},
	}
}

// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
// must be set even if it is a no-op function.
func (k *Keeper) SetUpgradeHandler(name string, upgradeHandler types.UpgradeHandler) {
	k.upgradeHandlers[name] = upgradeHandler
}

// ScheduleUpgrade schedules an upgrade based on the specified plan
func (k *Keeper) ScheduleUpgrade(ctx sdk.Context, plan types.Plan) sdk.Error {
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
	} else if plan.Height <= ctx.BlockHeight() {
		return sdk.ErrUnknownRequest("Upgrade cannot be scheduled in the past")
	}
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.DoneHeightKey(plan.Name)) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("Upgrade with name %s has already been completed", plan.Name))
	}
	bz := k.cdc.MustMarshalBinaryBare(plan)
	k.haveCache = false
	store.Set(types.PlanKey(), bz)
	return nil
}

// ClearUpgradePlan clears any schedule upgrade
func (k *Keeper) ClearUpgradePlan(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	k.haveCache = false
	store.Delete(types.PlanKey())
}

// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
// upgrade or false if there is none
func (k *Keeper) GetUpgradePlan(ctx sdk.Context) (plan types.Plan, havePlan bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PlanKey())
	if bz == nil {
		return plan, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &plan)
	return plan, true
}

// setDone marks this upgrade name as being done so the name can't be reused accidentally
func (k *Keeper) setDone(ctx sdk.Context, name string) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	store.Set(types.DoneHeightKey(name), bz)
}

// BeginBlocker should be called inside the BeginBlocker method of any app using the upgrade module. Scheduled upgrade
// plans are cached in memory so the overhead of this method is trivial.
func (k *Keeper) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) {
	blockTime := ctx.BlockHeader().Time
	blockHeight := ctx.BlockHeight()

	plan, found := k.GetUpgradePlan(ctx)
	if !found {
		return
	}

	upgradeTime := plan.Time
	upgradeHeight := plan.Height
	if (!upgradeTime.IsZero() && !blockTime.Before(upgradeTime)) || (upgradeHeight > 0 && upgradeHeight <= blockHeight) {
		handler, ok := k.upgradeHandlers[plan.Name]
		if ok {
			// We have an upgrade handler for this upgrade name, so apply the upgrade
			ctx.Logger().Info(fmt.Sprintf("Applying upgrade \"%s\" at height %d", plan.Name, blockHeight))
			handler(ctx, plan)
			k.ClearUpgradePlan(ctx)
			k.setDone(ctx, plan.Name)
		} else {
			// We don't have an upgrade handler for this upgrade name, meaning this software is out of date so shutdown
			ctx.Logger().Error(fmt.Sprintf("UPGRADE \"%s\" NEEDED at height %d: %s", plan.Name, blockHeight, plan.Info))
			panic("UPGRADE REQUIRED!")
		}
	} else {
		// if we have a pending upgrade, but it is not yet time, make sure we did not
		// set the handler already
		_, ok := k.upgradeHandlers[plan.Name]
		if ok {
			ctx.Logger().Error(fmt.Sprintf("UNKNOWN UPGRADE \"%s\" - in binary but not executed on chain", plan.Name))
			panic("BINARY UPDATED BEFORE TRIGGER!")
		}
	}
}
