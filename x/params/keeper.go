package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params/space"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc  *codec.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	spaces map[string]*Space
}

// NewKeeper construct a params keeper
func NewKeeper(cdc *codec.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey) (k Keeper) {
	k = Keeper{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		spaces: make(map[string]*Space),
	}

	return k
}

// Allocate substore used for keepers
func (k Keeper) Subspace(spacename string) Space {
	_, ok := k.spaces[spacename]
	if ok {
		panic("subspace already occupied")
	}

	if spacename == "" {
		panic("cannot use empty string for subspace")
	}

	space := space.NewSpace(k.cdc, k.key, k.tkey, spacename)

	k.spaces[spacename] = &space

	return space
}

// Get existing subspace from keeper
func (k Keeper) GetSubspace(spacename string) (Space, bool) {
	space, ok := k.spaces[spacename]
	if !ok {
		return Space{}, false
	}
	return *space, ok
}
