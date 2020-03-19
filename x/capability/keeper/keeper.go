package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

type (
	// Keeper defines the capability module's keeper. It is responsible for provisioning,
	// tracking, and authenticating capabilities at runtime. During application
	// initialization, the keeper can be hooked up to modules through unique function
	// references so that it can identify the calling module when later invoked.
	//
	// When the initial state is loaded from disk, the keeper allows the ability to
	// create new capability keys for all previously allocated capability identifiers
	// (allocated during execution of past transactions and assigned to particular modes),
	// and keep them in a memory-only store while the chain is running.
	//
	// The keeper allows the ability to create scoped sub-keepers which are tied to
	// a single specific module.
	Keeper struct {
		cdc           *codec.Codec
		storeKey      sdk.StoreKey
		memKey        sdk.StoreKey
		scopedModules map[string]struct{}
		sealed        bool
	}

	// ScopedKeeper defines a scoped sub-keeper which is tied to a single specific
	// module provisioned by the capability keeper. Scoped keepers must be created
	// at application initialization and passed to modules, which can then use them
	// to claim capabilities they receive and retrieve capabilities which they own
	// by name, in addition to creating new capabilities & authenticating capabilities
	// passed by other modules.
	ScopedKeeper struct {
		cdc      *codec.Codec
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey
		module   string
	}
)

func NewKeeper(cdc *codec.Codec, storeKey, memKey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		scopedModules: make(map[string]struct{}),
		sealed:        false,
	}
}

// ScopeToModule attempts to create and return a ScopedKeeper for a given module
// by name. It will panic if the keeper is already sealed or if the module name
// already has a ScopedKeeper.
func (k Keeper) ScopeToModule(moduleName string) ScopedKeeper {
	if k.sealed {
		panic("cannot scope to module via a sealed capability keeper")
	}

	if _, ok := k.scopedModules[moduleName]; ok {
		panic(fmt.Sprintf("cannot create multiple scoped keepers for the same module name: %s", moduleName))
	}

	k.scopedModules[moduleName] = struct{}{}

	return ScopedKeeper{
		storeKey: k.storeKey,
		memKey:   k.memKey,
		module:   moduleName,
	}
}

// AuthenticateCapability attempts to authenticate a given capability and name
// from a caller. It allows for a caller to check that a capability does in fact
// correspond to a particular name. The scoped keeper will lookup the capability
// from the internal in-memory store and check against the provided name. It returns
// true upon success and false upon failure.
//
// Note, the capability's forward mapping is indexed by a string which should
// contain it's unique memory reference.
func (sk ScopedKeeper) AuthenticateCapability(ctx sdk.Context, cap types.Capability, name string) bool {
	memStore := ctx.KVStore(sk.memKey)

	bz := memStore.Get(types.FwdCapabilityKey(sk.module, cap))
	return string(bz) == name
}

// ClaimCapability attempts to claim a given Capability and name tuple. This tuple
// is considered an Owner. It will attempt to add the owner to the persistent
// set of capability owners for the capability index. If the owner already exists,
// it will return an error. Otherwise, it will also set a forward and reverse index
// for the capability and capability name.
func (sk ScopedKeeper) ClaimCapability(ctx sdk.Context, cap types.Capability, name string) error {
	store := prefix.NewStore(ctx.KVStore(sk.storeKey), types.KeyPrefixIndexCapability)
	memStore := ctx.KVStore(sk.memKey)
	indexKey := types.IndexToKey(cap.GetIndex())

	var capOwners *types.CapabilityOwners

	bz := store.Get(indexKey)
	if len(bz) == 0 {
		capOwners = types.NewCapabilityOwners()
	} else {
		sk.cdc.MustUnmarshalBinaryBare(bz, capOwners)
	}

	if err := capOwners.Set(types.NewOwner(sk.module, name)); err != nil {
		return err
	}

	// update capability owner set
	store.Set(indexKey, sk.cdc.MustMarshalBinaryBare(capOwners))

	// Set the forward mapping between the module and capability tuple and the
	// capability name in the in-memory store.
	memStore.Set(types.FwdCapabilityKey(sk.module, cap), []byte(name))

	// Set the reverse mapping between the module and capability name and the
	// capability in the in-memory store.
	memStore.Set(types.RevCapabilityKey(sk.module, name), sk.cdc.MustMarshalBinaryBare(cap))

	return nil
}
