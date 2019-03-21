package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper - crisis keeper
type Keeper struct {
	routes     InvarRoutes
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
func (k *Keeper) RegisterRoute(route string, invar sdk.Invariant) {
	invarRoute := NewInvarRoute(route, invar)
	k.routes = append(k.routes, invarRoute)
}

// register routes for the
func (k *Keeper) Routes() Routes {
	return k.routes.Routes()
}
