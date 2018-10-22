package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc  sdk.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	spaces map[string]*Subspace
}

// NewKeeper constructs a params keeper
func NewKeeper(cdc sdk.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey) (k Keeper) {
	k = Keeper{
		cdc:  cdc.Cache(),
		key:  key,
		tkey: tkey,

		spaces: make(map[string]*Subspace),
	}

	return k
}

// Allocate subspace used for keepers
func (k Keeper) Subspace(spacename string) Subspace {
	_, ok := k.spaces[spacename]
	if ok {
		panic("subspace already occupied")
	}

	if spacename == "" {
		panic("cannot use empty string for subspace")
	}

	space := subspace.NewSubspace(k.cdc, k.key, k.tkey, spacename)

	k.spaces[spacename] = &space

	return space
}

// Get existing substore from keeper
func (k Keeper) GetSubspace(storename string) (Subspace, bool) {
	space, ok := k.spaces[storename]
	if !ok {
		return Subspace{}, false
	}
	return *space, ok
}
