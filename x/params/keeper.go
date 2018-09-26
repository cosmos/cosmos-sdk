package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params/store"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc  *codec.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	stores map[string]*Store
}

// NewKeeper constructs a params keeper
func NewKeeper(cdc *codec.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey) (k Keeper) {
	k = Keeper{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		stores: make(map[string]*Store),
	}

	return k
}

// Allocate substore used for keepers
func (k Keeper) Substore(storename string) Store {
	_, ok := k.stores[storename]
	if ok {
		panic("substore already occupied")
	}

	if storename == "" {
		panic("cannot use empty string for substore")
	}

	store := store.NewStore(k.cdc, k.key, k.tkey, storename)

	k.stores[storename] = &store

	return store
}

// Get existing substore from keeper
func (k Keeper) GetSubstore(storename string) (Store, bool) {
	store, ok := k.stores[storename]
	if !ok {
		return Store{}, false
	}
	return *store, ok
}
