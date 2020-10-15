package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/convo/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

// Keeper defines the IBC convo keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler

	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
}

// NewKeeper creates a new IBC convo Keeper instance
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey,
	channelKeeper types.ChannelKeeper, portKeeper types.PortKeeper, scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      key,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", host.ModuleName, types.ModuleName))
}

// SetPendingMessage sets the pending message for the sender-receiver convo over the giver
// source channel. It will replace any previously set pending message
func (k Keeper) SetPendingMessage(ctx sdk.Context, sender, srcChannel, receiver, msg string) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPendingMessage(sender, srcChannel, receiver)
	store.Set(key, []byte(msg))
}

// GetPendingMessage gets the pending message for the sender-receiver convo over the giver
// source channel.
func (k Keeper) GetPendingMessage(ctx sdk.Context, sender, srcChannel, receiver string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPendingMessage(sender, srcChannel, receiver)
	return string(store.Get(key))
}

// DeletePendingMessage deletes the pending message for the sender-receiver convo over the giver
// source channel.
func (k Keeper) DeletePendingMessage(ctx sdk.Context, sender, srcChannel, receiver string) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPendingMessage(sender, srcChannel, receiver)
	store.Delete(key)
}

// SetInboxMessage will set the incoming message intended for the receiver from the sender
// over the destination channel in the receiver's inbox (receiver has an inbox per channel)
func (k Keeper) SetInboxMessage(ctx sdk.Context, sender, srcChannel, receiver, msg string) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyInboxMessage(sender, srcChannel, receiver)
	store.Set(key, []byte(msg))
}

// GetInboxMessage will get the incoming message intended for the receiver from the sender
// over the destination channel in the receiver's inbox (receiver has an inbox per channel)
func (k Keeper) GetInboxMessage(ctx sdk.Context, sender, srcChannel, receiver string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyInboxMessage(sender, srcChannel, receiver)
	return string(store.Get(key))
}

// SetOutboxMessage will set the latest confirmed outgoing message of the sender-receiver convo over the giver
// source channel
func (k Keeper) SetOutboxMessage(ctx sdk.Context, sender, srcChannel, receiver, msg string) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyOutboxMessage(sender, srcChannel, receiver)
	store.Set(key, []byte(msg))
}

// GetOutboxMessage will get the latest confirmed outgoing message of the sender-receiver convo over the giver
// source channel
func (k Keeper) GetOutboxMessage(ctx sdk.Context, sender, srcChannel, receiver string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyOutboxMessage(sender, srcChannel, receiver)
	return string(store.Get(key))
}

// ChanCloseInit defines a wrapper function for the channel Keeper's function
// in order to expose it to the ICS20 transfer handler.
func (k Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	capName := host.ChannelCapabilityPath(portID, channelID)
	chanCap, ok := k.scopedKeeper.GetCapability(ctx, capName)
	if !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "could not retrieve channel capability at: %s", capName)
	}
	return k.channelKeeper.ChanCloseInit(ctx, portID, channelID, chanCap)
}

// IsBound checks if the transfer module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the ort Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort returns the portID for the transfer module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the transfer module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
