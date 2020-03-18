package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		memKey        sdk.MemoryStoreKey
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
		storeKey sdk.StoreKey
		memKey   sdk.MemoryStoreKey
		module   string
	}
)

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, memKey sdk.MemoryStoreKey) Keeper {
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
