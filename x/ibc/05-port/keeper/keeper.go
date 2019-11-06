package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store
	ports     map[sdk.CapabilityKey]string
	bound     []string
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/port",
		prefix:    []byte{},
		// prefix:    []byte(types.SubModuleName + "/"),                                          // "port/"
		ports: make(map[sdk.CapabilityKey]string), // map of capabilities to port ids
	}
}

// GetPorts returns the list of bound ports.
// (these ports still must have been bound statically)
func (k Keeper) GetPorts() []string {
	return k.bound
}

// GetPort retrieves a given port ID from keeper map
func (k Keeper) GetPort(ck sdk.CapabilityKey) (string, bool) {
	portID, found := k.ports[ck]
	return portID, found
}

// BindPort binds to a port and returns the associated capability.
// Ports must be bound statically when the chain starts in `app.go`.
// The capability must then be passed to a module which will need to pass
// it as an extra parameter when calling functions on the IBC module.
func (k Keeper) BindPort(portID string) sdk.CapabilityKey {
	if err := host.DefaultPortIdentifierValidator(portID); err != nil {
		panic(err.Error())
	}

	for _, b := range k.bound {
		if b == portID {
			panic(fmt.Sprintf("port %s is already bound", portID))
		}
	}

	key := sdk.NewKVStoreKey(portID)
	k.ports[key] = portID
	k.bound = append(k.bound, portID)
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

	return k.ports[key] == portID
}
