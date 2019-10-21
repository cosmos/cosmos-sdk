package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store
	ports     map[sdk.CapabilityKey]string
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/port",
		prefix:    []byte(types.SubModuleName + "/"),                                          // "port/"
	}
}

// BindPort binds to a port, returning the capability key which must be given to a module
func (k Keeper) BindPort(portID string) sdk.CapabilityKey {
	if !types.ValidatePortID(portID) {
		panic(fmt.Sprintf("invalid port id: %s", types.ErrInvalidPortID(k.codespace)))
	}

	key := sdk.NewKVStoreKey(portID)
	k.ports[key] = portID
	return key
}

// Authenticate authenticates a key against a port ID
func (k Keeper) Authenticate(key sdk.CapabilityKey, portID string) bool {
	return k.ports[key] == portID
}
