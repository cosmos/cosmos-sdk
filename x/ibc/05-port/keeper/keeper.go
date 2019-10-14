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
func (k Keeper) GetPort(ctx sdk.Context, portID string) (string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyPort(portID))
	if bz == nil {
		return "", false
	}

	return string(bz), true
}

// SetPort sets a port to the store
func (k Keeper) SetPort(ctx sdk.Context, portID string, capabilityKey sdk.CapabilityKey) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(capabilityKey)
	store.Set(types.KeyPort(portID), bz)
}

// BindPort binds to an unallocated port, failing if the port has already been
// allocated.
func (k Keeper) BindPort(ctx sdk.Context, portID string, generateFn exported.Generate) sdk.Error {
	_, found := k.GetPort(ctx, portID)
	if found {
		return types.ErrPortExists(k.codespace)
	}
	key := generateFn()
	k.SetPort(ctx, portID, key)
	return nil
}
