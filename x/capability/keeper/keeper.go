package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		cdc           codec.Marshaler
		storeKey      sdk.StoreKey
		capStore      types.CapabilityStore
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
		cdc      codec.Marshaler
		storeKey sdk.StoreKey
		capStore types.CapabilityStore // shared amongst all scoped keepers
		module   string
	}
)

func NewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		capStore:      types.NewCapabilityStore(),
		scopedModules: make(map[string]struct{}),
		sealed:        false,
	}
}

// ScopeToModule attempts to create and return a ScopedKeeper for a given module
// by name. It will panic if the keeper is already sealed or if the module name
// already has a ScopedKeeper.
func (k *Keeper) ScopeToModule(moduleName string) ScopedKeeper {
	if k.sealed {
		panic("cannot scope to module via a sealed capability keeper")
	}

	if _, ok := k.scopedModules[moduleName]; ok {
		panic(fmt.Sprintf("cannot create multiple scoped keepers for the same module name: %s", moduleName))
	}

	k.scopedModules[moduleName] = struct{}{}

	return ScopedKeeper{
		cdc:      k.cdc,
		storeKey: k.storeKey,
		capStore: k.capStore,
		module:   moduleName,
	}
}

// InitializeAndSeal loads all capabilities from the persistent KVStore into the
// in-memory store and seals the keeper to prevent further modules from creating
// a scoped keeper. InitializeAndSeal must be called once after the application
// state is loaded.
func (k *Keeper) InitializeAndSeal(ctx sdk.Context) {
	if k.sealed {
		panic("cannot initialize and seal an already sealed capability keeper")
	}

	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIndexCapability)
	iterator := sdk.KVStorePrefixIterator(prefixStore, nil)

	// initialize the in-memory store for all persisted capabilities
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		index := types.IndexFromKey(iterator.Key())
		cap := types.NewCapability(index)

		var capOwners types.CapabilityOwners
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &capOwners)

		for _, owner := range capOwners.Owners {
			// Set the forward mapping between the module and capability tuple and the
			// capability name in the in-memory store.
			k.capStore.SetCapabilityName(owner.Module, owner.Name, cap)

			// Set the reverse mapping between the module and capability name and the
			// capability in the in-memory store.
			k.capStore.SetCapability(owner.Module, owner.Name, cap)
		}
	}

	k.sealed = true
}

// GetLatestIndex returns the latest index of the CapabilityKeeper
func (k Keeper) GetLatestIndex(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	return types.IndexFromKey(store.Get(types.KeyIndex))
}

// NewCapability attempts to create a new capability with a given name. If the
// capability already exists in the in-memory store, an error will be returned.
// Otherwise, a new capability is created with the current global unique index.
// The newly created capability has the scoped module name and capability name
// tuple set as the initial owner. Finally, the global index is incremented along
// with forward and reverse indexes set in the in-memory store.
//
// Note, namespacing is completely local, which is safe since records are prefixed
// with the module name and no two ScopedKeeper can have the same module name.
func (sk ScopedKeeper) NewCapability(ctx sdk.Context, name string) (*types.Capability, error) {
	store := ctx.KVStore(sk.storeKey)

	if cap := sk.capStore.GetCapability(sk.module, name); cap != nil {
		return nil, sdkerrors.Wrapf(types.ErrCapabilityTaken, fmt.Sprintf("module: %s, name: %s", sk.module, name))

	}

	// create new capability with the current global index
	index := types.IndexFromKey(store.Get(types.KeyIndex))
	cap := types.NewCapability(index)

	// update capability owner set
	if err := sk.addOwner(ctx, cap, name); err != nil {
		return nil, err
	}

	// increment global index
	store.Set(types.KeyIndex, types.IndexToKey(index+1))

	// Set the forward mapping between the module and capability tuple and the
	// capability name in the in-memory store.
	sk.capStore.SetCapabilityName(sk.module, name, cap)

	// Set the reverse mapping between the module and capability name and the
	// capability in the in-memory store.
	sk.capStore.SetCapability(sk.module, name, cap)

	logger(ctx).Info("created new capability", "module", sk.module, "name", name)
	return cap, nil
}

// AuthenticateCapability attempts to authenticate a given capability and name
// from a caller. It allows for a caller to check that a capability does in fact
// correspond to a particular name. The scoped keeper will lookup the capability
// from the internal in-memory store and check against the provided name. It returns
// true upon success and false upon failure.
//
// Note, the capability's forward mapping is indexed by a string which should
// contain its unique memory reference.
func (sk ScopedKeeper) AuthenticateCapability(ctx sdk.Context, cap *types.Capability, name string) bool {
	return sk.capStore.GetCapabilityName(sk.module, cap) == name
}

// ClaimCapability attempts to claim a given Capability. The provided name and
// the scoped module's name tuple are treated as the owner. It will attempt
// to add the owner to the persistent set of capability owners for the capability
// index. If the owner already exists, it will return an error. Otherwise, it will
// also set a forward and reverse index for the capability and capability name.
func (sk ScopedKeeper) ClaimCapability(ctx sdk.Context, cap *types.Capability, name string) error {
	// update capability owner set
	if err := sk.addOwner(ctx, cap, name); err != nil {
		return err
	}

	// Set the forward mapping between the module and capability tuple and the
	// capability name in the in-memory store.
	sk.capStore.SetCapabilityName(sk.module, name, cap)

	// Set the reverse mapping between the module and capability name and the
	// capability in the in-memory store.
	sk.capStore.SetCapability(sk.module, name, cap)

	logger(ctx).Info("claimed capability", "module", sk.module, "name", name, "capability", cap.GetIndex())
	return nil
}

// ReleaseCapability allows a scoped module to release a capability which it had
// previously claimed or created. After releasing the capability, if no more
// owners exist, the capability will be globally removed.
func (sk ScopedKeeper) ReleaseCapability(ctx sdk.Context, cap *types.Capability) error {
	name := sk.capStore.GetCapabilityName(sk.module, cap)
	if len(name) == 0 {
		return sdkerrors.Wrap(types.ErrCapabilityNotOwned, sk.module)
	}

	// Remove the forward mapping between the module and capability tuple and the
	// capability name in the in-memory store.
	sk.capStore.DeleteCapabilityName(sk.module, cap)

	// Remove the reverse mapping between the module and capability name and the
	// capability in the in-memory store.
	sk.capStore.DeleteCapability(sk.module, name)

	// remove owner
	capOwners := sk.getOwners(ctx, cap)
	capOwners.Remove(types.NewOwner(sk.module, name))

	prefixStore := prefix.NewStore(ctx.KVStore(sk.storeKey), types.KeyPrefixIndexCapability)
	indexKey := types.IndexToKey(cap.GetIndex())

	if len(capOwners.Owners) == 0 {
		// remove capability owner set
		prefixStore.Delete(indexKey)
	} else {
		// update capability owner set
		prefixStore.Set(indexKey, sk.cdc.MustMarshalBinaryBare(capOwners))
	}

	return nil
}

// GetCapability allows a module to fetch a capability which it previously claimed
// by name. The module is not allowed to retrieve capabilities which it does not
// own.
func (sk ScopedKeeper) GetCapability(ctx sdk.Context, name string) (*types.Capability, bool) {
	cap := sk.capStore.GetCapability(sk.module, name)
	if cap == nil {
		return nil, false
	}

	return cap, true
}

// Get all the Owners that own the capability associated with the name this ScopedKeeper uses
// to refer to the capability
func (sk ScopedKeeper) GetOwners(ctx sdk.Context, name string) (*types.CapabilityOwners, bool) {
	cap, ok := sk.GetCapability(ctx, name)
	if !ok {
		return nil, false
	}

	prefixStore := prefix.NewStore(ctx.KVStore(sk.storeKey), types.KeyPrefixIndexCapability)
	indexKey := types.IndexToKey(cap.GetIndex())

	var capOwners types.CapabilityOwners

	bz := prefixStore.Get(indexKey)
	if len(bz) == 0 {
		return nil, false
	}

	sk.cdc.MustUnmarshalBinaryBare(bz, &capOwners)
	return &capOwners, true

}

// LookupModules returns all the module owners for a given capability
// as a string array, the capability is also returned along with a boolean success flag
func (sk ScopedKeeper) LookupModules(ctx sdk.Context, name string) ([]string, *types.Capability, bool) {
	cap, ok := sk.GetCapability(ctx, name)
	if !ok {
		return nil, nil, false
	}

	capOwners, ok := sk.GetOwners(ctx, name)
	if !ok {
		return nil, nil, false
	}

	mods := make([]string, len(capOwners.Owners))
	for i, co := range capOwners.Owners {
		mods[i] = co.Module
	}
	return mods, cap, true

}

func (sk ScopedKeeper) addOwner(ctx sdk.Context, cap *types.Capability, name string) error {
	prefixStore := prefix.NewStore(ctx.KVStore(sk.storeKey), types.KeyPrefixIndexCapability)
	indexKey := types.IndexToKey(cap.GetIndex())

	capOwners := sk.getOwners(ctx, cap)

	if err := capOwners.Set(types.NewOwner(sk.module, name)); err != nil {
		return err
	}

	// update capability owner set
	prefixStore.Set(indexKey, sk.cdc.MustMarshalBinaryBare(capOwners))
	return nil
}

func (sk ScopedKeeper) getOwners(ctx sdk.Context, cap *types.Capability) *types.CapabilityOwners {
	prefixStore := prefix.NewStore(ctx.KVStore(sk.storeKey), types.KeyPrefixIndexCapability)
	indexKey := types.IndexToKey(cap.GetIndex())

	bz := prefixStore.Get(indexKey)

	var owners *types.CapabilityOwners
	if len(bz) == 0 {
		owners = types.NewCapabilityOwners()
	} else {
		var capOwners types.CapabilityOwners
		sk.cdc.MustUnmarshalBinaryBare(bz, &capOwners)
		owners = &capOwners
	}

	return owners
}

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
