package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/params/subspace"
	"github.com/cosmos/cosmos-sdk/params/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc    codec.Marshaler
	key    sdk.StoreKey
	tkey   sdk.StoreKey
	spaces map[string]*subspace.Subspace
}

// NewKeeper constructs a params keeper
func NewKeeper(cdc codec.Marshaler, key, tkey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:    cdc,
		key:    key,
		tkey:   tkey,
		spaces: make(map[string]*subspace.Subspace),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Allocate subspace used for keepers
func (k Keeper) Subspace(s string) subspace.Subspace {
	_, ok := k.spaces[s]
	if ok {
		panic("subspace already occupied")
	}

	if s == "" {
		panic("cannot use empty string for subspace")
	}

	space := subspace.NewSubspace(k.cdc, k.key, k.tkey, s)
	k.spaces[s] = &space

	return space
}

// Get existing substore from keeper
func (k Keeper) GetSubspace(s string) (subspace.Subspace, bool) {
	space, ok := k.spaces[s]
	if !ok {
		return subspace.Subspace{}, false
	}
	return *space, ok
}
