package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/params/consensus"
	"github.com/cosmos/cosmos-sdk/x/params/gas"
	"github.com/cosmos/cosmos-sdk/x/params/msgstat"
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc  *wire.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	space KeeperSpace

	stores map[string]*Store
}

// KeeperSpace defines param space for the substores
// Zero value("") means not using the substore
type KeeperSpace struct {
	ConsensusSpace string
	GasConfigSpace string
	MsgStatusSpace string
}

// Default KeeperSpace
func DefaultKeeperSpace() *KeeperSpace {
	return &KeeperSpace{
		ConsensusSpace: consensus.DefaultParamSpace,
		GasConfigSpace: gas.DefaultParamSpace,
		MsgStatusSpace: msgstat.DefaultParamSpace,
	}
}

// NewKeeper construct a params keeper
func NewKeeper(cdc *wire.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey, space *KeeperSpace) (k Keeper) {
	if space == nil {
		space = DefaultKeeperSpace()
	}

	k = Keeper{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		space: *space,

		stores: make(map[string]*Store),
	}

	// Registering Default Subspaces
	if space != nil {
		if space.ConsensusSpace != "" {
			_ = k.SubStore(space.ConsensusSpace)
		}
		if space.GasConfigSpace != "" {
			_ = k.SubStore(space.GasConfigSpace)
		}
		if space.MsgStatusSpace != "" {
			_ = k.SubStore(space.MsgStatusSpace)
		}
	}

	return k
}

// Allocate substore used for keepers
func (k Keeper) SubStore(space string) Store {
	_, ok := k.stores[space]
	if ok {
		panic("substore already occupied")
	}

	return k.UnsafeSubStore(space)
}

// Get substore without checking existing allocation
func (k Keeper) UnsafeSubStore(space string) Store {
	if space == "" {
		panic("cannot use empty string for substore space")
	}

	store := store.NewStore(k.cdc, k.key, k.tkey, space)

	k.stores[space] = &store

	return store
}

// ConsensusStore returns ConsensusParams submodule store
// Panics if not allocated
func (k Keeper) ConsensusStore() Store {
	return *k.stores[k.space.ConsensusSpace]
}

// GasConfigStore returns GasConfig submodule store
// Panics if not allocated
func (k Keeper) GasConfigStore() Store {
	return *k.stores[k.space.GasConfigSpace]
}

// MsgStatusStore returns MsgStatus submodule store
// Panics if not allocated
func (k Keeper) MsgStatusStore() Store {
	return *k.stores[k.space.MsgStatusSpace]
}
