package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper - crisis keeper
type Keeper struct {
	routes     []InvarRoute
	paramSpace params.Subspace

	distrKeeper         DistrKeeper
	bankKeeper          BankKeeper
	feeCollectionKeeper FeeCollectionKeeper
}

// NewKeeper creates a new Keeper object
func NewKeeper(paramSpace params.Subspace,
	distrKeeper DistrKeeper, bankKeeper BankKeeper,
	feeCollectionKeeper FeeCollectionKeeper) Keeper {

	return Keeper{
		routes:              []InvarRoute{},
		paramSpace:          paramSpace.WithKeyTable(ParamKeyTable()),
		distrKeeper:         distrKeeper,
		bankKeeper:          bankKeeper,
		feeCollectionKeeper: feeCollectionKeeper,
	}
}

// register routes for the
func (k *Keeper) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	invarRoute := NewInvarRoute(moduleName, route, invar)
	k.routes = append(k.routes, invarRoute)
}

// Routes - return the keeper's invariant routes
func (k Keeper) Routes() []InvarRoute {
	return k.routes
}

// Invariants returns all the registered Crisis keeper invariants.
func (k Keeper) Invariants() []sdk.Invariant {
	var invars []sdk.Invariant
	for _, route := range k.routes {
		invars = append(invars, route.Invar)
	}
	return invars
}

// DONTCOVER
