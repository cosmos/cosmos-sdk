package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	scopedKeeper capability.ScopedKeeper
	cdc          *codec.Codec
	ports        map[string]bool
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(sck capability.ScopedKeeper, cdc *codec.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		scopedKeeper: sck,
		cdc:          cdc,
		ports:        make(map[string]bool), // map of capability key names to port ids
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
func (k *Keeper) BindPort(ctx sdk.Context, portID string) capability.Capability {
	if err := host.DefaultPortIdentifierValidator(portID); err != nil {
		panic(err.Error())
	}

	if k.isBounded(portID) {
		panic(fmt.Sprintf("port %s is already bound", portID))
	}

	key, err := k.scopedKeeper.NewCapability(ctx, portID)
	if err != nil {
		panic(err.Error())
	}
	k.ports[portID] = true

	return key
}

// Authenticate authenticates a capability key against a port ID
// by checking if the memory address of the capability was previously
// generated and bound to the port (provided as a parameter) which the capability
// is being authenticated against.
func (k Keeper) Authenticate(ctx sdk.Context, key capability.Capability, portID string) bool {
	if err := host.DefaultPortIdentifierValidator(portID); err != nil {
		panic(err.Error())
	}

	return k.scopedKeeper.AuthenticateCapability(ctx, key, portID)
}
