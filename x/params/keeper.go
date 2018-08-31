package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/params/space"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc  *wire.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	spaces map[string]*Space
}

// NewKeeper construct a params keeper
func NewKeeper(cdc *wire.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey) (k Keeper) {
	k = Keeper{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		spaces: make(map[string]*Space),
	}

	return k
}

// Allocate substore used for keepers
func (k Keeper) Subspace(space string) Space {
	_, ok := k.spaces[space]
	if ok {
		panic("subspace already occupied")
	}

	return k.UnsafeSubspace(space)
}

// Get substore without checking existing allocation
func (k Keeper) UnsafeSubspace(spacename string) Space {
	if spacename == "" {
		panic("cannot use empty string for subspace")
	}

	space := space.NewSpace(k.cdc, k.key, k.tkey, spacename)

	k.spaces[spacename] = &space

	return space
}
