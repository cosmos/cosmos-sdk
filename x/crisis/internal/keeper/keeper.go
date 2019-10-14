package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper - crisis keeper
type Keeper struct {
	routes         []types.InvarRoute
	paramSpace     params.Subspace
	invCheckPeriod uint

	supplyKeeper types.SupplyKeeper

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

// NewKeeper creates a new Keeper object
func NewKeeper(
	paramSpace params.Subspace, invCheckPeriod uint, supplyKeeper types.SupplyKeeper,
	feeCollectorName string,
) Keeper {

	return Keeper{
		routes:           make([]types.InvarRoute, 0),
		paramSpace:       paramSpace.WithKeyTable(types.ParamKeyTable()),
		invCheckPeriod:   invCheckPeriod,
		supplyKeeper:     supplyKeeper,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// RegisterRoute register the routes for each of the invariants
func (k *Keeper) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	invarRoute := types.NewInvarRoute(moduleName, route, invar)
	k.routes = append(k.routes, invarRoute)
}

// Routes - return the keeper's invariant routes
func (k Keeper) Routes() []types.InvarRoute {
	return k.routes
}

// Invariants returns all the registered Crisis keeper invariants.
func (k Keeper) Invariants() []sdk.Invariant {
	invars := make([]sdk.Invariant, len(k.routes))
	for i, route := range k.routes {
		invars[i] = route.Invar
	}
	return invars
}

// AssertInvariants asserts all registered invariants. If any invariant fails,
// the method panics.
func (k Keeper) AssertInvariants(ctx sdk.Context) {
	logger := k.Logger(ctx)

	start := time.Now()
	invarRoutes := k.Routes()

	for _, ir := range invarRoutes {
		if res, stop := ir.Invar(ctx); stop {
			// TODO: Include app name as part of context to allow for this to be
			// variable.
			panic(fmt.Errorf("invariant broken: %s\n"+
				"\tCRITICAL please submit the following transaction:\n"+
				"\t\t tx crisis invariant-broken %s %s", res, ir.ModuleName, ir.Route))
		}
	}

	end := time.Now()
	diff := end.Sub(start)

	logger.Info("asserted all invariants", "duration", diff, "height", ctx.BlockHeight())
}

// InvCheckPeriod returns the invariant checks period.
func (k Keeper) InvCheckPeriod() uint { return k.invCheckPeriod }

// SendCoinsFromAccountToFeeCollector transfers amt to the fee collector account.
func (k Keeper) SendCoinsFromAccountToFeeCollector(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return k.supplyKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, k.feeCollectorName, amt)
}
