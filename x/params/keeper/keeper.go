package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types/subspace"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc    codec.Marshaler
	key    sdk.StoreKey
	tkey   sdk.StoreKey
	spaces map[string]*params.Subspace
}

// NewKeeper constructs a params keeper
func NewKeeper(cdc codec.Marshaler, key, tkey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:    cdc,
		key:    key,
		tkey:   tkey,
		spaces: make(map[string]*params.Subspace),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", params.ModuleName))
}

// Allocate subspace used for keepers
func (k Keeper) Subspace(s string) params.Subspace {
	_, ok := k.spaces[s]
	if ok {
		panic("subspace already occupied")
	}

	if s == "" {
		panic("cannot use empty string for subspace")
	}

	space := params.NewSubspace(k.cdc, k.key, k.tkey, s)
	k.spaces[s] = &space

	return space
}

// Get existing substore from keeper
func (k Keeper) GetSubspace(s string) (params.Subspace, bool) {
	space, ok := k.spaces[s]
	if !ok {
		return params.Subspace{}, false
	}
	return *space, ok
}
