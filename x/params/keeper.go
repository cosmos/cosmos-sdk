package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc       *codec.Codec
	key       sdk.StoreKey
	tkey      sdk.StoreKey
	codespace sdk.CodespaceType
	spaces    map[string]*Subspace
}

// NewKeeper constructs a params keeper
func NewKeeper(cdc *codec.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey, codespace sdk.CodespaceType) (k Keeper) {
	k = Keeper{
		cdc:       cdc,
		key:       key,
		tkey:      tkey,
		codespace: codespace,
		spaces:    make(map[string]*Subspace),
	}

	return k
}

// Allocate subspace used for keepers
func (k Keeper) Subspace(s string) Subspace {
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
func (k Keeper) GetSubspace(s string) (Subspace, bool) {
	space, ok := k.spaces[s]
	if !ok {
		return Subspace{}, false
	}
	return *space, ok
}
