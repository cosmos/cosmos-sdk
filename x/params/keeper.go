package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/params/consensus"
	"github.com/cosmos/cosmos-sdk/x/params/gas"
	"github.com/cosmos/cosmos-sdk/x/params/msgstat"
	"github.com/cosmos/cosmos-sdk/x/params/space"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc  *wire.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	space KeeperSpace

	spaces map[string]*Space
}

// KeeperSpace defines param space for the subspaces
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

		spaces: make(map[string]*Space),
	}

	// Registering Default Subspaces
	if space != nil {
		if space.ConsensusSpace != "" {
			_ = k.Subspace(space.ConsensusSpace)
		}
		if space.GasConfigSpace != "" {
			_ = k.Subspace(space.GasConfigSpace)
		}
		if space.MsgStatusSpace != "" {
			_ = k.Subspace(space.MsgStatusSpace)
		}
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

// ConsensusSpace returns ConsensusParams submodule store
// Panics if not allocated
func (k Keeper) ConsensusSpace() Space {
	return *k.spaces[k.space.ConsensusSpace]
}

// GasConfigSpace returns GasConfig submodule store
// Panics if not allocated
func (k Keeper) GasConfigSpace() Space {
	return *k.spaces[k.space.GasConfigSpace]
}

// MsgStatusSpace returns MsgStatus submodule store
// Panics if not allocated
func (k Keeper) MsgStatusSpace() Space {
	return *k.spaces[k.space.MsgStatusSpace]
}
