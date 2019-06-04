package upgrade

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"os"
	"os/exec"
	"path/filepath"
)

// Keeper of the upgrade module
type Keeper interface {
	// ScheduleUpgrade schedules an upgrade based on the specified plan
	ScheduleUpgrade(ctx sdk.Context, plan Plan) sdk.Error

	// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
	// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
	// must be set even if it is a no-op function.
	SetUpgradeHandler(name string, upgradeHandler Handler)

	// ClearUpgradePlan clears any schedule upgrade
	ClearUpgradePlan(ctx sdk.Context)

	// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
	// upgrade or false if there is none
	GetUpgradePlan(ctx sdk.Context) (plan Plan, havePlan bool)

	// SetWillUpgrader sets a custom function to be run whenever an upgrade is scheduled. This
	// can be used to notify the node that an upgrade will be happen in the future so that it
	// can download any software ahead of time in the background.
	// It does not indicate that an upgrade is happening now and should just be used for preparation,
	// not the actual upgrade.
	SetWillUpgrader(willUpgrader func(ctx sdk.Context, plan Plan))

	// SetOnUpgrader sets a custom function to be called right before the chain halts and the
	// upgrade needs to be applied. This can be used to initiate an automatic upgrade process.
	SetOnUpgrader(onUpgrader func(ctx sdk.Context, plan Plan))

	// BeginBlocker should be called inside the BeginBlocker method of any app using the upgrade module. Scheduled upgrade
	// plans are cached in memory so the overhead of this method is trivial.
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock)
}

type keeper struct {
	storeKey        sdk.StoreKey
	cdc             *codec.Codec
	upgradeHandlers map[string]Handler
	haveCache       bool
	//haveCachedPlan  bool
	//plan            Plan
	willUpgrader    func(ctx sdk.Context, plan Plan)
	onUpgrader      func(ctx sdk.Context, plan Plan)
}

const (
	// PlanKey specifies the key under which an upgrade plan is stored in the store
	PlanKey = "plan"
)

// NewKeeper constructs an upgrade keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return &keeper{
		storeKey:        storeKey,
		cdc:             cdc,
		upgradeHandlers: map[string]Handler{},
	}
}

func (keeper *keeper) SetUpgradeHandler(name string, upgradeHandler Handler) {
	keeper.upgradeHandlers[name] = upgradeHandler
}

func (keeper *keeper) SetWillUpgrader(willUpgrader func(ctx sdk.Context, plan Plan)) {
	keeper.willUpgrader = willUpgrader
}

func (keeper *keeper) SetOnUpgrader(onUpgrader func(ctx sdk.Context, plan Plan)) {
	keeper.onUpgrader = onUpgrader
}

func (keeper *keeper) ScheduleUpgrade(ctx sdk.Context, plan Plan) sdk.Error {
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
	if store.Has(DoneHeightKey(plan.Name)) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("Upgrade with name %s has already been completed", plan.Name))
	}
	bz := keeper.cdc.MustMarshalBinaryBare(plan)
	keeper.haveCache = false
	store.Set([]byte(PlanKey), bz)
	return nil
}

func (keeper *keeper) ClearUpgradePlan(ctx sdk.Context) {
	store := ctx.KVStore(keeper.storeKey)
	keeper.haveCache = false
	store.Delete([]byte(PlanKey))
}

// ValidateBasic does basic validation of a Plan
func (plan Plan) ValidateBasic() sdk.Error {
	if len(plan.Name) == 0 {
		return sdk.ErrUnknownRequest("Name cannot be empty")

	}
	return nil
}

func (keeper *keeper) GetUpgradePlan(ctx sdk.Context) (plan Plan, havePlan bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte(PlanKey))
	if bz == nil {
		return plan, false
	}
	keeper.cdc.MustUnmarshalBinaryBare(bz, &plan)
	return plan, true
}

// DoneHeightKey returns a key that points to the height at which a past upgrade was applied
func DoneHeightKey(name string) []byte {
	return []byte(fmt.Sprintf("done/%s", name))
}

func (keeper *keeper) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) {
	blockTime := ctx.BlockHeader().Time
	blockHeight := ctx.BlockHeight()

	plan, found := keeper.GetUpgradePlan(ctx)

	if !found {
		return
	}

	upgradeTime := plan.Time
	upgradeHeight := plan.Height
	if (!upgradeTime.IsZero() && !blockTime.Before(upgradeTime)) || (upgradeHeight > 0 && upgradeHeight <= blockHeight) {
		handler, ok := keeper.upgradeHandlers[plan.Name]
		if ok {
			// We have an upgrade handler for this upgrade name, so apply the upgrade
			ctx.Logger().Info(fmt.Sprintf("Applying upgrade \"%s\" at height %d", plan.Name, blockHeight))
			handler(ctx, plan)
			keeper.ClearUpgradePlan(ctx)
			// Mark this upgrade name as being done so the name can't be reused accidentally
			store := ctx.KVStore(keeper.storeKey)
			bz := keeper.cdc.MustMarshalBinaryBare(blockHeight)
			store.Set(DoneHeightKey(plan.Name), bz)
		} else {
			// We don't have an upgrade handler for this upgrade name, meaning this software is out of date so shutdown
			ctx.Logger().Error(fmt.Sprintf("UPGRADE \"%s\" NEEDED at height %d: %s", plan.Name, blockHeight, plan.Info))
			if keeper.onUpgrader != nil {
				keeper.onUpgrader(ctx, plan)
			} else {
				DefaultOnUpgrader(ctx, plan)
			}
			panic("UPGRADE REQUIRED!")
		}
	}
	// TODO will upgraded is disabled for now because of inconsistent behavior
	//else if justRetrievedFromCache {
	//	// In this case we are notifying the system that an upgrade is planned but not scheduled to happen yet
	//	if keeper.willUpgrader != nil {
	//		keeper.willUpgrader(ctx, keeper.plan)
	//	} else {
	//		DefaultWillUpgrader(ctx, keeper.plan)
	//	}
	//}
}


// DefaultWillUpgrader asynchronously runs a script called prepare-upgrade from $COSMOS_HOME/config if such a script exists,
// with plan serialized to JSON as the first argument and the current block height as the second argument.
// The environment variable $COSMOS_HOME will be set to the home directory of the daemon.
func DefaultWillUpgrader(ctx sdk.Context, plan Plan) {
	CallUpgradeScript(ctx, plan, "prepare-upgrade", true)
}

// DefaultOnUpgrader synchronously runs a script called do-upgrade from $COSMOS_HOME/config if such a script exists,
// with plan serialized to JSON as the first argument and the current block height as the second argument.
// The environment variable $COSMOS_HOME will be set to the home directory of the daemon.
func DefaultOnUpgrader(ctx sdk.Context, plan Plan) {
	CallUpgradeScript(ctx, plan, "do-upgrade", false)
}

// CallUpgradeScript runs a script called script from $COSMOS_HOME/config if such a script exists,
// with plan serialized to JSON as the first argument and the current block height as the second argument.
// The environment variable $COSMOS_HOME will be set to the home directory of the daemon.
// If async is true, the command will be run in a separate go-routine.
func CallUpgradeScript(ctx sdk.Context, plan Plan, script string, async bool) {
	f := func() {
		home := viper.GetString(cli.HomeFlag)
		file := filepath.Join(home, "config", script)
		ctx.Logger().Info(fmt.Sprintf("Looking for upgrade script %s", file))
		if _, err := os.Stat(file); err == nil {
			ctx.Logger().Info(fmt.Sprintf("Applying upgrade script %s", file))
			err = os.Setenv("COSMOS_HOME", home)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error setting env var COSMOS_HOME: %v", err))
			}


			planJson, err := json.Marshal(plan)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error marshaling upgrade plan to JSON: %v", err))
			}
			cmd := exec.Command(file, string(planJson), fmt.Sprintf("%d", ctx.BlockHeight()))
			cmd.Stdout = logWriter{ctx, script, false}
			cmd.Stderr = logWriter{ctx, script, false}
			err = cmd.Run()
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error running script %s: %v", file, err))
			}
		}
	}
	if async {
		go f()
	} else {
		f()
	}
}

type logWriter struct {
	sdk.Context
	script string
	err    bool
}

func (w logWriter) Write(p []byte) (n int, err error) {
	s := fmt.Sprintf("script %s: %s", w.script, p)
	if w.err {
		w.Logger().Error(s)
	} else {
		w.Logger().Info(s)
	}
	return len(p), nil
}
