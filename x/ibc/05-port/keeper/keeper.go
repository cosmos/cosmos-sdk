package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/05-port/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/ports",
		prefix:    []byte(types.SubModuleName + "/"),                                          // "ports/"
	}
}

// GetPort returns a port with a particular identifier
func (k Keeper) GetPort(ctx sdk.Context, portID string) (sdk.CapabilityKey, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyPort(portID))
	if bz == nil {
		return nil, false
	}

	var capabilityKey sdk.CapabilityKey
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &capabilityKey)
	return capabilityKey, true
}

// SetPort sets a port to the store
func (k Keeper) SetPort(ctx sdk.Context, portID string, capabilityKey sdk.CapabilityKey) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(capabilityKey)
	store.Set(types.KeyPort(portID), bz)
}

// delete a port ID key from the store
func (k Keeper) deletePort(ctx sdk.Context, portID string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Delete(types.KeyPort(portID))
}

// BindPort binds to an unallocated port, failing if the port has already been
// allocated.
func (k Keeper) BindPort(ctx sdk.Context, portID string, generateFn exported.Generate) sdk.Error {
	if !types.ValidatePortID(portID) {
		return types.ErrInvalidPortID(k.codespace)
	}

	_, found := k.GetPort(ctx, portID)
	if found {
		return types.ErrPortExists(k.codespace)
	}
	key := generateFn()
	k.SetPort(ctx, portID, key)
	return nil
}

// TransferPort allows an existing port (i.e capability key) to be transfered
// to another module and thus stored under another path.
//
// NOTE: not neccessary if the host state machine supports object-capabilities,
// since the port reference is a bearer capability.
func (k Keeper) TransferPort(
	ctx sdk.Context,
	portID string,
	authenticateFn exported.Authenticate,
	generateFn exported.Generate,
) sdk.Error {
	port, found := k.GetPort(ctx, portID)
	if !found {
		return types.ErrPortNotFound(k.codespace)
	}

	if !authenticateFn(port) {
		return types.ErrPortNotAuthenticated(k.codespace)
	}

	key := generateFn()
	k.SetPort(ctx, portID, key)
	return nil
}

// ReleasePort allows a module to release a port such that other modules may
// then bind to it.
//
// WARNING: releasing a port will allow other modules to bind to that port and
// possibly intercept incoming channel opening handshakes. Modules should release
// ports only when doing so is safe.
func (k Keeper) ReleasePort(
	ctx sdk.Context,
	portID string,
	authenticateFn exported.Authenticate) sdk.Error {
	port, found := k.GetPort(ctx, portID)
	if !found {
		return types.ErrPortNotFound(k.codespace)
	}

	if !authenticateFn(port) {
		return types.ErrPortNotAuthenticated(k.codespace)
	}

	k.deletePort(ctx, portID)
	return nil
}
