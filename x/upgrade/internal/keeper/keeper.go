package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/upgrade/internal/types"
)

type Keeper struct {
	skipUpgradeHeights map[int64]bool
	storeKey           sdk.StoreKey
	cdc                *codec.Codec
	upgradeHandlers    map[string]types.UpgradeHandler
}

// NewKeeper constructs an upgrade Keeper
func NewKeeper(skipUpgradeHeights map[int64]bool, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		skipUpgradeHeights: skipUpgradeHeights,
		storeKey:           storeKey,
		cdc:                cdc,
		upgradeHandlers:    map[string]types.UpgradeHandler{},
	}
}

// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
// must be set even if it is a no-op function.
func (k Keeper) SetUpgradeHandler(name string, upgradeHandler types.UpgradeHandler) {
	k.upgradeHandlers[name] = upgradeHandler
}

// ScheduleUpgrade schedules an upgrade based on the specified plan.
// If there is another Plan already scheduled, it will overwrite it
// (implicitly cancelling the current plan)
func (k Keeper) ScheduleUpgrade(ctx sdk.Context, plan types.Plan) error {
	if err := plan.ValidateBasic(); err != nil {
		return err
	}

	if !plan.Time.IsZero() {
		if !plan.Time.After(ctx.BlockHeader().Time) {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "upgrade cannot be scheduled in the past")
		}
	} else if plan.Height <= ctx.BlockHeight() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "upgrade cannot be scheduled in the past")
	}

	if k.GetDoneHeight(ctx, plan.Name) != 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "upgrade with name %s has already been completed", plan.Name)
	}

	bz := k.cdc.MustMarshalBinaryBare(plan)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PlanKey(), bz)

	return nil
}

// GetDoneHeight returns the height at which the given upgrade was executed
func (k Keeper) GetDoneHeight(ctx sdk.Context, name string) int64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{types.DoneByte})
	bz := store.Get([]byte(name))
	if len(bz) == 0 {
		return 0
	}

	return int64(binary.BigEndian.Uint64(bz))
}

// ClearUpgradePlan clears any schedule upgrade
func (k Keeper) ClearUpgradePlan(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.PlanKey())
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
// upgrade or false if there is none
func (k Keeper) GetUpgradePlan(ctx sdk.Context) (plan types.Plan, havePlan bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PlanKey())
	if bz == nil {
		return plan, false
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &plan)
	return plan, true
}

// setDone marks this upgrade name as being done so the name can't be reused accidentally
func (k Keeper) setDone(ctx sdk.Context, name string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{types.DoneByte})
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	store.Set([]byte(name), bz)
}

// HasHandler returns true iff there is a handler registered for this name
func (k Keeper) HasHandler(name string) bool {
	_, ok := k.upgradeHandlers[name]
	return ok
}

// ApplyUpgrade will execute the handler associated with the Plan and mark the plan as done.
func (k Keeper) ApplyUpgrade(ctx sdk.Context, plan types.Plan) {
	handler := k.upgradeHandlers[plan.Name]
	if handler == nil {
		panic("ApplyUpgrade should never be called without first checking HasHandler")
	}

	handler(ctx, plan)

	k.ClearUpgradePlan(ctx)
	k.setDone(ctx, plan.Name)
}

// IsSkipHeight checks if the given height is part of skipUpgradeHeights
func (k Keeper) IsSkipHeight(height int64) bool {
	return k.skipUpgradeHeights[height]
}
