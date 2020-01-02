package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	ports    map[string]bool
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
		ports:    make(map[string]bool), // map of capability key names to port ids
	}
}

// isBounded checks a given port ID is already bounded.
func (k Keeper) isBounded(portID string) bool {
	return k.ports[portID]
}

// BindPort binds to a port and returns the associated capability.
// Ports must be bound statically when the chain starts in `app.go`.
// The capability must then be passed to a module which will need to pass
// it as an extra parameter when calling functions on the IBC module.
func (k *Keeper) BindPort(portID string) sdk.CapabilityKey {
	if err := host.DefaultPortIdentifierValidator(portID); err != nil {
		panic(err.Error())
	}

	if k.isBounded(portID) {
		panic(fmt.Sprintf("port %s is already bound", portID))
	}

	key := sdk.NewKVStoreKey(portID)
	k.ports[key.Name()] = true // NOTE: key name and value always match

	return key
}

// Authenticate authenticates a capability key against a port ID
// by checking if the memory address of the capability was previously
// generated and bound to the port (provided as a parameter) which the capability
// is being authenticated against.
func (k Keeper) Authenticate(key sdk.CapabilityKey, portID string) bool {
	if err := host.DefaultPortIdentifierValidator(portID); err != nil {
		panic(err.Error())
	}

	if key.Name() != portID {
		return false
	}

	return k.ports[key.Name()]
}
